-- Placeholder live warp check. A full Warpra destination list + locked-buy
-- refusal is Phase 7 (uses this harness). This script only confirms the
-- character is in the world, then quits.
local done = false

function tick()
	if done then
		goro.quit()
		return
	end
	local p = goro.player()
	if p.x ~= nil then
		done = true
	end
end
