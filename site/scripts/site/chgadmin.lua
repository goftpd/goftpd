if #params ~= 2 then
	session:Reply(501, "Syntax: site chgadmin <user> <group>")
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
	if not u:HasGroup(params[2]) then
		err = Error()
		err.Message = "User isn't part of '" .. params[2] .. "'"
		return err
	end

	if u.Groups[params[2]:lower()].IsAdmin then
		session:Reply(226, "Removed gadmin permission for " .. target.Name .. " in " .. params[2])
	else
		session:Reply(226, "Added gadmin permission for " .. target.Name .. " in " .. params[2])
	end

	u.Groups[params[2]:lower()].IsAdmin = not u.Groups[params[2]:lower()].IsAdmin

	return nil
end)

if err then
	session:Reply(500, "Error: " .. err:Error())
	return false
end

return true
