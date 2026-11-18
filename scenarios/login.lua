-- Headless login smoke for Phase 5.2. kgap pool characters only.
-- Usage: goro --headless --script scenarios/login.lua
local walked = false

function tick()
	if walked then
		goro.quit()
		return
	end
	local p = goro.player()
	if p.x ~= nil then
		goro.walk(p.x + 1, p.y)
		walked = true
	end
end
