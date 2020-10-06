if #params < 3 then
	session:Reply(501, "Syntax: site change <user> <field> <setting>")
	return false
end

-- check if we have it
local target, err = session:Auth():GetUser(params[1])
if err then
	session:Reply(500, "Error: " .. err:Error())
	return false
end

local caller = session:User()

-- check that the caller is allowed
if not acl:MatchTarget(caller, target) then
	session:Reply(500, "Permission denied")
	return false
end

local field = params[2]

local err = session:Auth():UpdateUser(target.Name, function(u)
	if field == "ratio" then
		u.Ratio = tonumber(params[3])

	else
		err = Error()
		err.Message = "Unknown field"
		return err
	end

	return nil
end)
if err then
	session:Reply(500, "Error: " .. err:Error())
	return false
end

session:Reply(226, "Updated user!")
return true
