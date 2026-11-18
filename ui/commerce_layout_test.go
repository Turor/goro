package ui

import (
	"testing"

	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestCommerceWindowsFollowLayoutReload(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	manager := NewManager()
	sess := session.New()
	sess.Inventory.Items = []session.InventoryItem{{
		Index: 1, ItemID: 2301, Amount: 1, Type: db.ItemTypeArmor, Identified: false,
	}}
	ctx := Context{
		ScreenW: 1280, ScreenH: 800,
		UIManager: manager,
		Input:     input.NewState(),
		Session:   sess,
	}

	var shop ShopWindow
	shop.OpenBuy(nil, ctx)
	var deal ShopWindow
	deal.OpenDeal(network.ShopDealSelection{NPCID: 1}, ctx)
	var trade TradeWindow
	trade.Open(ctx, "Tester")
	var vending VendingWindow
	vending.OpenBuy(ctx, network.VendingItemList{})
	var cart CartWindow
	cart.OpenWindow(ctx)
	var storage StorageWindow
	storage.OpenWindow(ctx)
	var mail MailWindow
	mail.Open(ctx, nil)
	mail.ShowMail(network.MailMessage{ID: 1, Title: "Hello", Sender: "Tester", Body: "Hi"})
	var changeCart ChangeCartWindow
	changeCart.Open(ctx)
	var prompt amountPrompt
	prompt.Open(ctx, "Input number", 1, 10, nil)
	var teleport TeleportModal
	teleport.OpenWarpPointList(network.WarpPointList{SkillID: teleportSkillID, MapNames: []string{"prontera"}})
	if !teleport.Update(ctx) || teleport.content == nil {
		t.Fatal("teleport window did not open")
	}
	var identify IdentifyWindow
	identify.OpenList(ctx, network.ItemIdentifyList{Indexes: []uint16{1}})
	var cards CardCompositionWindow
	cards.OpenList(ctx, 2, network.ItemCompositionList{Indexes: []uint16{1}})
	var arrows MakingArrowWindow
	arrows.OpenList(ctx, network.MakingArrowList{ItemIDs: []uint32{1750}})
	var making MakingItemWindow
	making.OpenList(ctx, network.MakingItemList{Items: []network.MakingItemOption{{ItemID: 501}}})
	var repair RepairItemWindow
	repair.OpenList(ctx, network.RepairItemList{Items: []network.RepairItem{{Index: 1, ItemID: 2301}}})
	var refine WeaponRefineWindow
	refine.OpenList(ctx, network.WeaponRefineList{Items: []network.WeaponRefineItem{{Index: 1, ItemID: 1101}}})
	var egg PetEggWindow
	egg.OpenList(ctx, network.PetEggList{Indexes: []uint16{1}})
	var spell AutoSpellWindow
	spell.OpenList(ctx, network.AutoSpellList{SkillIDs: []uint16{11}}, nil)

	windows := []*Window{
		&shop.buyWindow, &shop.buyCartWindow, &deal.dealWindow, &trade.Window, &vending.leftWindow, &vending.rightWindow,
		&cart.Window, &storage.Window, &mail.Window, &mail.readWindow, &changeCart.Window, &prompt.Window,
		&teleport.Window, &identify.Window, &cards.Window, &arrows.Window, &making.Window,
		&repair.Window, &refine.Window, &egg.Window, &spell.Window,
	}
	for i, w := range windows {
		if !w.IsOpen() || w.content == nil {
			t.Fatalf("window %d did not stay open", i)
		}
	}
	shopContent := shop.buyWindow.content
	dealContent := deal.dealWindow.content
	defaultShopW := shop.buyWindow.width
	defaultCartW := cart.width
	defaultQty := shopQtyColor
	if defaultShopW != 420 || defaultCartW != cartGridLeftPad+cartGridCols*cartGridCell+ROScrollbarGutter {
		t.Fatalf("default shop width %d cart width %d", defaultShopW, defaultCartW)
	}

	set := func(section, key, raw string) {
		t.Helper()
		if err := rotheme.SetToken(section, key, raw); err != nil {
			t.Fatalf("set [%s] %s=%s: %v", section, key, raw, err)
		}
	}
	set("shop", "buy_list_width", "510")
	set("shop", "quantity_color", "#ff00ff")
	set("trade", "width", "640")
	set("vending", "width", "480")
	set("cart", "cols", "8")
	set("storage", "table_width", "400")
	set("mail", "width", "520")
	set("mail", "read_width", "480")
	set("change_cart", "width", "300")
	set("amount_prompt", "width", "260")
	set("teleport", "width", "320")
	set("teleport", "status_color", "#00ff00")
	set("identify", "width", "360")
	set("card_composition", "width", "400")
	set("making_arrow", "width", "360")
	set("making_item", "width", "360")
	set("equipment_choice", "width", "360")
	set("pet_egg", "width", "360")
	set("auto_spell", "width", "360")

	if shop.buyWindow.width != defaultShopW || cart.width != defaultCartW {
		t.Fatal("window size changed before rebuild")
	}
	if n := manager.RebuildStaleWindows(); n != len(windows) {
		t.Fatalf("rebuilt %d windows, want %d", n, len(windows))
	}
	if shop.buyWindow.width != 510 || shop.buyWindow.content == shopContent {
		t.Fatalf("shop buy width=%d content rebuilt=%v", shop.buyWindow.width, shop.buyWindow.content != shopContent)
	}
	if deal.dealWindow.width != smallPromptWidth || deal.dealWindow.content == dealContent {
		t.Fatalf("shop deal width=%d content rebuilt=%v", deal.dealWindow.width, deal.dealWindow.content != dealContent)
	}
	if shopQtyColor != widget.RGBA8(255, 0, 255, 255) {
		t.Fatalf("quantity color = %#v", shopQtyColor)
	}
	if teleportStatusColor != widget.RGBA8(0, 255, 0, 255) {
		t.Fatalf("teleport status color = %#v", teleportStatusColor)
	}
	checks := []struct {
		name      string
		got, want int
	}{
		{"trade", trade.width, 640},
		{"vending", vending.leftWindow.width, 480},
		{"vending cart", vending.rightWindow.width, 480},
		{"cart", cart.width, cartWindowWidth},
		{"storage", storage.width, storageWindowWidth},
		{"mail", mail.width, 520},
		{"mail read", mail.readWindow.width, 480},
		{"change cart", changeCart.width, 300},
		{"amount", prompt.width, 260},
		{"teleport", teleport.width, 320},
		{"identify", identify.width, 360},
		{"cards", cards.width, 400},
		{"arrows", arrows.width, 360},
		{"making", making.width, 360},
		{"repair", repair.width, 360},
		{"refine", refine.width, 360},
		{"egg", egg.width, 360},
		{"spell", spell.width, 360},
		{"shop cart", shop.buyCartWindow.width, shopBuyCartWindowW},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Fatalf("%s width = %d, want %d", check.name, check.got, check.want)
		}
	}
	if cart.width != cartGridLeftPad+8*cartGridCell+ROScrollbarGutter {
		t.Fatalf("cart width = %d after cols=8", cart.width)
	}

	rotheme.ResetLayout()
	if n := manager.RebuildStaleWindows(); n != len(windows) {
		t.Fatalf("restore rebuilt %d windows, want %d", n, len(windows))
	}
	if shop.buyWindow.width != defaultShopW || cart.width != defaultCartW || trade.width != 560 || teleport.width != 260 {
		t.Fatalf("restored shop=%d cart=%d trade=%d teleport=%d", shop.buyWindow.width, cart.width, trade.width, teleport.width)
	}
	if shopQtyColor != defaultQty || teleportStatusColor != widget.RGBA8(204, 48, 48, 255) {
		t.Fatalf("colors quantity=%#v status=%#v", shopQtyColor, teleportStatusColor)
	}
}
