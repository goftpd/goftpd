if #params ~= 1 then
	session:Reply(501, "Syntax: site user <user>")
	return false
end

local user, err = session:Auth():GetUser(params[1])
if err then
	session:Reply(500, "Error: " .. err:Error())
	return false
end

session:Reply(226, "User: " .. user.Name)

for i, v in user.IPMasks() do
	session:Reply(226, "Mask [" .. i .. "]: " .. v)
end

return true
