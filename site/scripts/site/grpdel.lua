-- site addip <user> <mask> <...mask>
if #params ~= 1 then
	session:Reply(501, "Syntax: site grpass <group>")
	return false
end

-- check if we have it
local target, err = session:Auth():GetGroup(params[1])
if err then
	session:Reply(500, "Group '" .. params[1] .. "' doesn't exist")
	return false
end

-- attempt to add the group
err = session:Auth():DeleteGroup(params[1])
if err then
	session:Reply(500, "Error: " .. err:Error())
	return false
end

session:Reply(226, "Deleted group '" .. params[1] .."'")

return true

