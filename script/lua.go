package script

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/goftpd/goftpd/acl"
	"github.com/goftpd/goftpd/ftp/cmd"
	"github.com/pkg/errors"
	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
	"golang.org/x/sync/errgroup"
	luar "layeh.com/gopher-luar"
)

type ScriptError struct {
	Message string
}

func (e ScriptError) Error() string { return e.Message }

type LUAEngine struct {
	// compiled lua byte code is stored here
	byteCode map[string]*lua.FunctionProto

	// need to organise our commands for easy access to paths
	// i.e. map[FTPCommand][]Command
	commands map[string][]Command

	// pool of lstate would be nice, but no Reset
}

func NewLUAEngine(lines []string) (*LUAEngine, error) {
	le := &LUAEngine{
		byteCode: make(map[string]*lua.FunctionProto, 0),
		commands: make(map[string][]Command, 0),
	}

	for _, l := range lines {
		reader := csv.NewReader(strings.NewReader(l))
		reader.Comma = ' '

		fields, err := reader.Read()
		if err != nil {
			return nil, err
		}

		if len(fields) < 5 {
			return nil, errors.New("expected at least 5 fields")
		}

		hook, ok := stringToScriptHook[fields[0]]
		if !ok {
			return nil, errors.Errorf("unexpected script hook '%s'", fields[0])
		}

		stype, ok := stringToScriptType[fields[2]]
		if !ok {
			return nil, errors.Errorf("unexpected script type '%s'", fields[2])
		}

		a, err := acl.NewFromString(strings.Join(fields[4:], " "))
		if err != nil {
			return nil, errors.Errorf("unable to parse acl '%s': %s", strings.Join(fields[4:], " "), err)
		}

		c := Command{
			FTPCommand: strings.ToLower(fields[1]),
			Path:       fields[3],
			Hook:       hook,
			ScriptType: stype,
			ACL:        a,
		}

		if _, ok := le.commands[c.FTPCommand]; !ok {
			le.commands[c.FTPCommand] = make([]Command, 0)
		}

		le.commands[c.FTPCommand] = append(le.commands[c.FTPCommand], c)
	}

	for _, all := range le.commands {
		for _, c := range all {
			if err := le.compileFile(c.Path); err != nil {
				return nil, errors.WithMessage(err, c.Path)
			}
		}
	}

	return le, nil
}

// compileFile opens a path and attempts to compile it and store
// the byte code for reuse
func (le *LUAEngine) compileFile(path string) error {
	if _, ok := le.byteCode[path]; ok {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	chunk, err := parse.Parse(reader, path)
	if err != nil {
		return err
	}

	proto, err := lua.Compile(chunk, path)
	if err != nil {
		return err
	}

	le.byteCode[path] = proto

	return nil
}

// Do takes in a context path to the script and a cmd.Session and tries to execute the
// script
func (le *LUAEngine) Do(pctx context.Context, fields []string, hook ScriptHook, session cmd.Session) error {
	ftpCommand := strings.ToLower(fields[0])

	// TODO:
	// this is slightly inefficient as we do a lot of allocating here around fields
	// before we check to see if we actually have this

	if ftpCommand == "site" {
		if len(fields) < 2 {
			return nil
		}
		ftpCommand = strings.ToLower(strings.Join(fields[0:2], " "))
		fields = fields[2:]
	} else {

		if len(fields) > 1 {
			fields = fields[1:]
		} else {
			fields = []string{}
		}
	}

	if _, ok := le.commands[ftpCommand]; !ok {
		return ErrNotExist
	}

	// TODO
	// wrap context with a deadline
	errg, ctx := errgroup.WithContext(pctx)

	for _, c := range le.commands[ftpCommand] {

		// check Trigger/Hook/ScriptType
		if c.Hook != hook {
			continue
		}

		proto, ok := le.byteCode[c.Path]
		if !ok {
			return errors.New("script not found")
		}

		// check permissions
		user := session.User()
		if user == nil {
			return errors.New("user is nil")
		}

		if m, ok := c.ACL.ExplicitMatch(user); !m && ok {
			session.ReplyStatus(cmd.StatusPermissionDenied)
			return ErrStop
		}

		// IMPORTANT
		// you have to also check MatchTarget to check for self and gadmin actions
		// but script is responsible for this

		fn := func(ctx context.Context) func() error {
			return func() error {
				L := lua.NewState()
				defer L.Close()

				// TODO: do we need to use context as it degrades performance quite a lot
				// although we could cancel all the concurrent scripts with it also :/
				L.SetContext(ctx)

				// push byte code
				L.Push(L.NewFunctionFromProto(proto))

				L.SetGlobal("ftpCommand", luar.New(L, ftpCommand))
				L.SetGlobal("Error", luar.NewType(L, ScriptError{}))
				L.SetGlobal("params", luar.New(L, fields))
				L.SetGlobal("session", luar.New(L, session))
				L.SetGlobal("acl", luar.New(L, c.ACL))

				if err := L.PCall(0, 1, nil); err != nil {
					return err
				}

				ret := L.Get(-1)
				L.Pop(1)

				if ret.Type() != lua.LTBool {
					return errors.Errorf("expected bool in return to %s", c.Path)
				}

				// if false dont continue, aka return an error
				if !lua.LVAsBool(ret) {
					return ErrStop
				}

				return nil
			}
		}

		if c.ScriptType == ScriptTypeEvent {
			// add a deadline and then call
			ctx, _ := context.WithTimeout(pctx, time.Second*60)
			go func(ctx context.Context, ftpCommand, path string) {
				if err := fn(ctx)(); err != nil {
					fmt.Fprintf(os.Stderr, "ERROR event '%s' '%s': %s", ftpCommand, path, err)
				}
			}(ctx, ftpCommand, c.Path)

		} else {
			errg.Go(fn(ctx))
		}
	}

	if err := errg.Wait(); err != nil {
		return err
	}

	return nil
}
