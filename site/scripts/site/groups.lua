local groups, err = session:Auth():GetGroups()
if err then
	session:Reply(500, "Error: " .. err:Error())
	return false
end

if not groups then
	session:Reply(226, "No groups found")
	return true
end

session:Reply(226, "Found " .. #groups .. " groups:")

if groups then
	for i, group in groups() do
		session:Reply(226, " " .. group.Name .. " - " .. group.Description)
	end
end


return true
