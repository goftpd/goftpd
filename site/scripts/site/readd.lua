local caller = session:User()

if #params == 0 then
	-- get list of users that are deleted
	local users, err = session:Auth():GetUsers()
	if err then
		session:Reply(500, "Error getting users: " .. err:Error())
		return false
	end

	if #users == 0 then
		session:Reply(226, "No deleted users")
		return true
	end

	session:Reply(226, "Deleted users:")

	for i, target in users() do
		if not target.DeletedAt:IsZero() then
			-- check we have permission to view this deleted user
			if acl:MatchTarget(caller, target) then
				session:Reply(226, " " .. target.Name)
			end
		end
	end 

	return true
end

for i, user in params() do

	local target, err = session:Auth():GetUser(user)

	-- check we have permission to readd this user
	if not acl:MatchTarget(caller, target) then
		session:Reply(550, "Not allowed to readd '" .. user .. "'")
	else

		-- readd user
		local err = session:Auth():UpdateUser(target.Name, function(u)
			u:Readd()
			return nil
		end)

		if err then
			session:Reply(500, "Error readding user '" .. target.Name .. "': " .. err:Error())
			return false
		end

		session:Reply(226, "Readded '" .. target.Name .. "'")
	end
end

return true
