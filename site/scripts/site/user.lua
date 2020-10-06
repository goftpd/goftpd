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
session:Reply(226, "Credits: " .. target.Credits / 1024 .. "MB")
session:Reply(226, "Ratio: 1:" .. target.Ratio)
session:Reply(226, "Added By: " .. target.AddedBy)
session:Reply(226, "Created: " .. target.CreatedAt:Format("15:04 02/01/2006"))
session:Reply(226, "Last Login: " .. target.LastLoginAt:Format("15:04 02/01/2006"))
-- pretty redundant as last login at sets this also
-- session:Reply(226, "Last Updated: " .. target.UpdatedAt:Format("15:04 02/01/2006"))

if target.IPMasks then
	for i, mask in target.IPMasks() do
		session:Reply(226, "Mask [" .. i .. "]: " .. mask)
	end
end

if target.PrimaryGroup ~= "" then
	session:Reply(226, "Primary Group: " .. target.PrimaryGroup)
end

if target.Groups then
	for group, settings in target.Groups() do
		if settings.IsAdmin then
			session:Reply(226, "Group *" .. group)
		else
			session:Reply(226, "Group " .. group)
		end
	end
end

return true
