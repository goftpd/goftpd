if #params ~= 1 then
	session:Reply(501, "Syntax: site user <user>")
	return false
end

local caller = session:User()

-- ignore err as we dont want to leak if the user exists 
-- or not to the caller and MatchTarget checks for nil
local target, err = session:Auth():GetUser(params[1])

if not acl:MatchTarget(caller, target) then
	session:Reply(500, "Permission denied")
	return false
end

session:Reply(226, "User: " .. target.Name)

if target.IPMasks ~= nil then
	for i, v in target.IPMasks() do
		session:Reply(226, "Mask [" .. i .. "]: " .. v)
	end
end

return true
