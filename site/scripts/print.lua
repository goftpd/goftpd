local pzsng = require('site/scripts/pzsng')

local user, ok = session:User()
if not ok then
	-- return false to prevent continuation and cancel current contexts?
	return false
end

pzsng.log("HI FROM " .. user.Name .. " cwd: " .. session:CWD())

if command then
	pzsng.log("Command: " .. command)
end

if params then
	for i, p in params() do
		pzsng.log("param " .. p)
	end
end

session:Reply(226, "HI FROM LUA!")

return true
