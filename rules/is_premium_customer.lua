function check(data)
	-- The engine can pass arbitrary data.
	local customer_type = data.customer_type
	return customer_type == "premium"
end
