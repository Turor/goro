package network

// ClientDate20250716 is the TuroranRO packet version: Hercules PACKETVER
// 20250716 (main client, renewal). It sits next to the upstream 20080910
// profile; the profile is chosen with --packet-client-date.
const ClientDate20250716 = 20250716

// IsClientDate20250716 reports whether a client date selects the 20250716
// wire profile (char info 175 bytes, 0x0AC4 login, 0x09FF actors, ...).
func IsClientDate20250716(clientDate int) bool {
	return clientDate >= ClientDate20250716
}

// PacketLengthsForClientDate returns the server-to-client length table for a
// packet profile.
func PacketLengthsForClientDate(clientDate int) LengthTable {
	if IsClientDate20250716(clientDate) {
		return PacketLengths20250716()
	}
	return PacketLengths2008()
}

// PacketLengths20250716 is the Hercules 20250716 table plus every Turoran
// custom packet length.
func PacketLengths20250716() LengthTable {
	table := make(LengthTable, len(packetLengths20250716Hercules)+len(turoranCustomLengths))
	for id, length := range packetLengths20250716Hercules {
		table[id] = length
	}
	for id, length := range turoranCustomLengths {
		table[id] = length
	}
	return table
}

// turoranCustomLengths are server-to-client packets added by the Turoran
// Hercules plugins (hercules-prere-plus src/plugins/turoran_*.c). They share
// the map stream with Hercules packets, so an unknown length desyncs the
// framer. New custom packets keep the 0x0EEx/0x0EFx range and are added here
// in the same PR as the plugin. 0x0EF1 is also a client-to-server opcode (the
// stable homunculus request); both directions use length 6.
var turoranCustomLengths = LengthTable{
	0x0EE8: -1, // ZC_ITEM_MARKS_LIST (turoran_item_marks), opt-in
	0x0EE9: 7,  // ZC_ITEM_MARK_RESULT (turoran_item_marks)
	0x0EF1: 6,  // ZC_TURORAN_FEATURES (turoran_item_marks), every client after every map load
	0x0EF5: -1, // ZC_MVP_MINIMAP (turoran_mvp_minimap), every client on MVP maps
	0x0EF6: -1, // stable owned materials (turoran_stable)
	0x0EF8: -1, // stable homunculus list (turoran_stable)
	0x0EF9: -1, // stable pet list (turoran_stable)
	0x0EFA: 8,  // stable result (turoran_stable)
	0x0EFB: -1, // Eden quest list (turoran_eden)
	0x0EFC: 8,  // Eden result (turoran_eden)
	0x0EFD: -1, // Eden shop (turoran_eden)
	0x0EFE: -1, // Warpra destination list (turoran_warpra)
	0x0EFF: 8,  // Warpra result (turoran_warpra)
}
