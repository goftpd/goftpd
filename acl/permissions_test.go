package acl

import (
	"fmt"
	"testing"
)

func TestPermissions(t *testing.T) {
	type input struct {
		line string
		want error
	}

	newInput := func(scope, acl string, want error) input {
		return input{
			line: fmt.Sprintf("%s %s", scope, acl),
			want: want,
		}
	}

	type test struct {
		inputs        []input
		matchUser     *User
		matchPath     string
		want          bool
		wantNoDefault bool
		wantMatch     bool
	}

	for scope := range StringToPermissionScope {
		tests := map[string]test{
			"basic validation": test{
				inputs: []input{
					newInput(scope, "/**", ErrRuleInvalidInput),
				},
			},
			"unknown scope": test{
				inputs: []input{
					newInput("downloaad", "/[* *", ErrRuleUnknownPermissionScope),
				},
			},
			"bad acl": test{
				inputs: []input{
					newInput(scope, "/* -@$", ErrACLInvalidCharacters),
				},
			},
			"glob compilation": test{
				inputs: []input{
					newInput(scope, "/[* *", ErrRuleBadGlob),
				},
			},
			"any user matches": test{
				inputs: []input{
					newInput(scope, "/** *", nil),
				},
				matchUser:     newUser("alice", "users"),
				matchPath:     "/goftpd",
				want:          true,
				wantNoDefault: true,
				wantMatch:     true,
			},
			"any user does not match": test{
				inputs: []input{
					newInput(scope, "/** !*", nil),
				},
				matchUser:     newUser("alice", "users"),
				matchPath:     "/goftpd",
				want:          false,
				wantNoDefault: false,
				wantMatch:     true,
			},
			"super user always matches": test{
				inputs: []input{
					newInput(scope, "/** !*", nil),
				},
				matchUser:     SuperUser,
				matchPath:     "/goftpd",
				want:          true,
				wantNoDefault: true,
				wantMatch:     true,
			},
			"any user matches on nested rule": test{
				inputs: []input{
					newInput(scope, "/** !*", nil),
					newInput(scope, "/dir/* *", nil),
				},
				matchUser:     newUser("alice", "users"),
				matchPath:     "/dir/goftpd",
				want:          true,
				wantNoDefault: true,
				wantMatch:     true,
			},
			"user matches on nested rule": test{
				inputs: []input{
					newInput(scope, "/** !*", nil),
					newInput(scope, "/dir/* -alice", nil),
				},
				matchUser:     newUser("alice", "users"),
				matchPath:     "/dir/goftpd",
				want:          true,
				wantNoDefault: true,
				wantMatch:     true,
			},
			"group matches on nested rule": test{
				inputs: []input{
					newInput(scope, "/** !*", nil),
					newInput(scope, "/dir/* =users", nil),
				},
				matchUser:     newUser("alice", "users"),
				matchPath:     "/dir/goftpd",
				want:          true,
				wantNoDefault: true,
				wantMatch:     true,
			},
			"group does not match when user is prohibited": test{
				inputs: []input{
					newInput(scope, "/** !*", nil),
					newInput(scope, "/dir/* !-alice =users", nil),
				},
				matchUser:     newUser("alice", "users"),
				matchPath:     "/dir/goftpd",
				want:          false,
				wantNoDefault: false,
				wantMatch:     true,
			},
			"user does not match when group is prohibited": test{
				inputs: []input{
					newInput(scope, "/** !*", nil),
					newInput(scope, "/dir/* -alice !=users", nil),
				},
				matchUser:     newUser("alice", "users"),
				matchPath:     "/dir/goftpd",
				want:          false,
				wantNoDefault: false,
				wantMatch:     true,
			},
			"no rule": test{
				inputs:        []input{},
				matchUser:     newUser("alice", "users"),
				matchPath:     "/goftpd",
				want:          false,
				wantNoDefault: false,
				wantMatch:     false,
			},
			"no explicit match": test{
				inputs: []input{
					newInput(scope, "/** -bob", nil),
				},
				matchUser:     newUser("alice", "users"),
				matchPath:     "/goftpd",
				want:          false,
				wantNoDefault: false,
				wantMatch:     false,
			},
			"no path match": test{
				inputs: []input{
					newInput(scope, "/foo/** *", nil),
				},
				matchUser:     newUser("alice", "users"),
				matchPath:     "/goftpd",
				want:          false,
				wantNoDefault: false,
				wantMatch:     false,
			},
		}

		for name, tc := range tests {
			// run a sub test so we dont fail all of our tests
			t.Run(fmt.Sprintf("%s: %s", scope, name), func(t *testing.T) {

				// create our rules from our inputs
				var rules []Rule

				for _, i := range tc.inputs {
					rule, err := NewRule(i.line)
					if err != i.want {
						t.Fatalf("expected input error: %#v, got: %#v", i.want, err)
					}

					// if this is a failing error, stop
					if i.want != nil {
						return
					}

					rules = append(rules, rule)
				}

				// create permission and match it with our inputs
				permissions := NewPermissions(rules)

				result := permissions.Match(StringToPermissionScope[scope], tc.matchPath, tc.matchUser)
				if result != tc.want {
					t.Fatalf("expected result: %#v, got: %#v", tc.want, result)
				}

				resultNoDefault, match := permissions.MatchNoDefault(StringToPermissionScope[scope], tc.matchPath, tc.matchUser)
				if resultNoDefault != tc.wantNoDefault {
					t.Fatalf("expected resultNoDefault: %#v, got: %#v", tc.wantNoDefault, resultNoDefault)
				}

				if match != tc.wantMatch {
					t.Fatalf("expected match: %#v, got: %#v", tc.wantMatch, match)
				}
			})
		}
	}
}
