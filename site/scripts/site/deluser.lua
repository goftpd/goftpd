-- check we have params
if not params then
	session:Reply(501, "Syntax: site deluser <user>")
	return false
end

-- site addip <user> <mask> <...mask>
if #params ~= 1 then
	session:Reply(501, "Syntax: site deluser <user>")
	return false
end

local caller = session:User()
local target, err = session:Auth():GetUser(params[1])

-- check permissions
if not acl:MatchTarget(caller, target) then
	session:Reply(500, "Permission denied")
	return false
end

-- update the user 
local err = session:Auth():UpdateUser(target.Name, function(u)
	u:Delete()
	return nil
end)

if err ~= nil then
	session:Reply(500, "Error: " .. err:Error())
	return false
end

session:Reply(226, "Deleted user '" .. target.Name .. "'")

return true
