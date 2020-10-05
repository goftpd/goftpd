if #params ~= 2 then
	session:Reply(501, "Syntax: site chpgrp <user> <group>")
	return false
end

local caller = session:User()

local target, err = session:Auth():GetUser(params[1])
if err then
	session:Reply(500, "User does not exist")
	return false
end

local group, err = session:Auth():GetGroup(params[2])
if err then
	session:Reply(500, "Group does not exist")
	return false
end

err = session:Auth():UpdateUser(target.Name, function (u)
	return session:Auth():UpdateGroup(params[2], function(g)
		if g:AddUser(caller.Name, target.Name) then
			u:AddGroup(params[2])
			u.PrimaryGroup = params[2]
		else 
			err = Error()
			err.Message = "No slots available to add user to " .. name
			return err
		end

		return nil
	end)
end)

if err then
	session:Reply(500, "Error: " .. err:Error())
	return false
end

session:Reply(226, "Primary Group set to " .. params[2])

return true
