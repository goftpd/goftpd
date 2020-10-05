if #params < 3 then
	session:Reply(501, "Syntax: site grpchange <group> <field> <setting>")
	return false
end

-- check if we have it
local target, err = session:Auth():GetGroup(params[1])
if err then
	session:Reply(500, "Error: " .. err:Error())
	return false
end

local caller = session:User()

-- check that the caller is allowed
if not acl:MatchTargetGroup(caller, target) then
	session:Reply(500, "Permission denied")
	return false
end

local field = params[2]

local err = session:Auth():UpdateGroup(target.Name, function(g)
	if field == "slots" then
		g.Slots = tonumber(params[3])

	elseif field == "leech_slots" then
		g.LeechSlots = tonumber(params[3])

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

session:Reply(226, "Updated group!")
return true
