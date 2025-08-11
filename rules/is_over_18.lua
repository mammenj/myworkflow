-- This function is the entry point for the Go engine.
-- It receives a table (the workflow data) and returns a boolean.
function check(data)
	-- 'data' is the table passed from Go.
	-- Access fields using dot notation or square brackets.
	local age = data.age
	if age and age >= 18 then
		return true
	else
		return false
	end
end
