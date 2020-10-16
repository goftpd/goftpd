-- The post_check script receives 3 parameters from glftpd:
-- $1 - the name of the uploaded file
-- $2 - the directory the file was uploaded to
-- $3 - the CRC code of the file (with calc_crc enabled, else 00000000)
--      Note: if the uploader times out, the CRC code is 00000000 as well,
--      this to prevent that if one reconnected and started uploading the
--      same file again but left a stalled upload session alive, a bad crc
--      would be passed and the file would be deleted.

-- import lua lib filepath for some nice helpers
local filepath = require("filepath")

local path = session:FS():Join(session:CWD(), params)
local absolutepath = session:FS():JoinRoot(session:CWD(), params)

-- required for crc, possibly speed also
local entry, err = session:FS():GetEntry(path)
if err then
	session:Reply(500, "Error in GetEntry: " .. err:Error())
	return true
end

local user = session:User()

-- Usage: site/bin/zipscript-c <absolute filepath> <crc> <user> <group> <tagline> <speed> <section>
local cmd = string.format('site/bin/zipscript-c "%s" %s "%s" "%s" "tagline" 10000 "section"', absolutepath, entry:CRCHex(), user.Name, user.PrimaryGroup)

print(cmd)

local ret = os.execute(cmd)

print(ret)

if ret > 0 then
	session:Reply(500, "Error in post_check")
	return false
end

return true
