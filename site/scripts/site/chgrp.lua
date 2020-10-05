if #params < 2 then
	session:Reply(501, "Syntax: site chgrp <user> <group> [<group..N>]")
	return false
end

local caller = session:User()

local target, err = session:Auth():GetUser(params[1])
if err then
	session:Reply(500, "User does not exist")
	return false
end

err = session:Auth():UpdateUser(target.Name, function (u)
	for i, name in params() do
		if i > 1 then
			local group, err = session:Auth():GetGroup(name)

			if err then
				session:Reply(500, "Group '" .. name .. "' does not exist.")
			else 

				-- if the user exists then toggle it off
				if u:HasGroup(name) then
					local err = session:Auth():UpdateGroup(name, function(g)
						group:RemoveUser(target.Name)
						return nil
					end)
					if err then
						session:Reply(500, "Error removing user from group " .. name .. ": " .. err:Error())
						return err
					end

					u:RemoveGroup(name)
					session:Reply(226, "Removed from '" .. name .. "'")
				else 

					-- try and update group, any errors are probably due to no slots
					local err = session:Auth():UpdateGroup(name, function(g)
						if group:AddUser(caller.Name, target.Name) then
							u:AddGroup(name)
							session:Reply(226, "Added to '" .. name .. "'")
						else
							err = Error()
							err.Message = "No slots available to add user to " .. name
							return err

						end
						return nil
					end)
					if err then
						return err
					end
				end
			end
		end
	end

	return nil
end)

if err then
	session:Reply(500, "Error: " .. err:Error())
	return false
end

return true
