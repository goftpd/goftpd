local pzsng = {}

function pzsng.log(line)
	-- open a file for appending, 'a' creates the file if it doesnt exist
	file = io.open("pzsng.log", "a")

	-- sets the default output file
	io.output(file)

	io.write(line)
	io.write("\n")

	io.close(file)
end

return pzsng
