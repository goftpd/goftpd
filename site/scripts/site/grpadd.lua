-- site addip <user> <mask> <...mask>
if #params < 1 then
	session:Reply(501, "Syntax: site grpass <group> [description]")
	return false
end

-- check if we have it
local target, err = session:Auth():GetGroup(params[1])
if not err then
	session:Reply(500, "Group '" .. params[1] .. "' already exists")
	return false
end


-- attempt to add the group
target, err = session:Auth():AddGroup(params[1])
if err then
	session:Reply(500, "Error: " .. err:Error())
	return false
end

-- if we have a description, lets add it
if #params > 1 then
	local err = session:Auth():UpdateGroup(params[1], function(g)
		-- join params 2: in to a string
		for i, s in params() do
			if i > 1 then
				g.Description = g.Description .. s .. " "
			end
		end
		return nil
	end)

	if err then
		session:Reply(500, "Error: " .. err:Error())
		return false
	end
end

session:Reply(226, "Added group '" .. params[1] .."'")

return true
