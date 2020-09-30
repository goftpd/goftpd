-- check we have params
if not params then
	session:Reply(501, "Syntax: site delip <user> <mask> <...mask>")
	return false
end

-- site delip <user> <mask> <...mask>
if #params < 2 then
	session:Reply(501, "Syntax: site delip <user> <mask> <...mask>")
	return false
end

-- get current user
local user, ok = session:User()
if not ok then
	session:Reply(530, "Not logged in.")
	return false
end

-- get target user
local target, err = session:Auth():GetUser(params[1])
if err then
	session:Reply(500, "Error: " .. err:Error())
	return false
end

-- TODO
-- validate that this user is allowed to do delip for target

-- update the user deling each mask
local err = session:Auth():UpdateUser(target.Name, function(u)
	for i, mask in params() do
		if i > 1 then
			-- TODO
			-- if this casts to a number, delete by index
			local ok = u:DelIP(mask)
			if ok then
				session:Reply(226, "Deleted: " .. mask)
			end
		end
	end

	return nil
end)

if err ~= nil then
	session:Reply(500, "Error: " .. err:Error())
	return false
end

return true
