package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"backend/models"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ---------------------------------------------------------------------------
// Seed data
// ---------------------------------------------------------------------------

var userNames = []struct{ username, first, last string }{
	{"alex_hunter", "Alex", "Hunter"},
	{"sarah_m", "Sarah", "Mitchell"},
	{"jake_torres", "Jake", "Torres"},
	{"emma_vance", "Emma", "Vance"},
	{"ryan_cole", "Ryan", "Cole"},
	{"olivia_park", "Olivia", "Park"},
	{"noah_west", "Noah", "West"},
	{"chloe_reed", "Chloe", "Reed"},
	{"ethan_shaw", "Ethan", "Shaw"},
	{"ava_brooks", "Ava", "Brooks"},
	{"liam_foster", "Liam", "Foster"},
	{"mia_grant", "Mia", "Grant"},
	{"mason_hayes", "Mason", "Hayes"},
	{"sophia_lin", "Sophia", "Lin"},
	{"lucas_ford", "Lucas", "Ford"},
	{"isabella_wu", "Isabella", "Wu"},
	{"aiden_price", "Aiden", "Price"},
	{"ella_morgan", "Ella", "Morgan"},
	{"jackson_bell", "Jackson", "Bell"},
	{"grace_kim", "Grace", "Kim"},
}

type auctionTemplate struct {
	title       string
	description string
	imageURL    string
	minPrice    int64
	maxPrice    int64
	increment   int64
}

var auctionTemplates = []auctionTemplate{
	// Electronics
	{
		"Apple MacBook Pro 16\" M3 Max",
		"Barely used MacBook Pro with M3 Max chip, 36GB RAM, 1TB SSD. AppleCare+ until 2026. Comes with original box and accessories.",
		"https://images.unsplash.com/photo-1517336714731-489689fd1ca8?auto=format&fit=crop&w=800&q=80",
		2500, 4200, 50,
	},
	{
		"iPhone 15 Pro Max 256GB Natural Titanium",
		"Purchased 3 months ago. Immaculate condition, always used with case and screen protector. Includes original accessories and box.",
		"https://images.unsplash.com/photo-1592750652612-8e5748d97adf?auto=format&fit=crop&w=800&q=80",
		900, 1400, 25,
	},
	{
		"Sony Alpha A7R V Mirrorless Camera Body",
		"Professional full-frame mirrorless camera, 61MP sensor. Used for one commercial shoot. All original packaging included.",
		"https://images.unsplash.com/photo-1502920917128-1aa500764649?auto=format&fit=crop&w=800&q=80",
		3000, 4800, 50,
	},
	{
		"Sony WH-1000XM5 Wireless Headphones",
		"Industry-leading noise cancellation. Purchased new 6 months ago. Comes with all accessories and carry case.",
		"https://images.unsplash.com/photo-1505740420928-5e560c06d30e?auto=format&fit=crop&w=800&q=80",
		250, 420, 10,
	},
	{
		"Apple iPad Pro 12.9\" M2 WiFi + Cellular",
		"512GB Space Gray. Used lightly for digital art. Includes Apple Pencil 2nd gen and Magic Keyboard folio.",
		"https://images.unsplash.com/photo-1544244015-0df4b3ffc6b0?auto=format&fit=crop&w=800&q=80",
		900, 1500, 25,
	},
	{
		"Custom Gaming PC – RTX 4090 / i9-13900K",
		"High-end gaming rig: Intel i9-13900K, 64GB DDR5, RTX 4090 24GB, 2TB NVMe. Custom watercooling loop. Selling to upgrade to workstation.",
		"https://images.unsplash.com/photo-1587202372634-32705e3bf49c?auto=format&fit=crop&w=800&q=80",
		2800, 4500, 50,
	},
	{
		"Apple Watch Ultra 2 – Black Titanium",
		"49mm case, Alpine Loop band size M. Purchased 4 months ago. All original accessories and box included.",
		"https://images.unsplash.com/photo-1551816230-ef5deaed4a26?auto=format&fit=crop&w=800&q=80",
		650, 950, 25,
	},
	{
		"DJI Mavic 3 Pro Fly More Combo",
		"Tri-lens drone with Hasselblad camera. 5 flights total, no crashes. Includes 3 batteries, carrying bag, and all accessories.",
		"https://images.unsplash.com/photo-1521405924368-64c5b84bec60?auto=format&fit=crop&w=800&q=80",
		1800, 2800, 50,
	},
	{
		"Samsung 65\" Neo QLED 4K Smart TV QN90C",
		"2023 model with mini-LED backlighting. Excellent picture quality. Used for 8 months in a smoke-free home.",
		"https://images.unsplash.com/photo-1593359677879-a4021073889b?auto=format&fit=crop&w=800&q=80",
		1200, 2000, 50,
	},
	{
		"Meta Quest 3 512GB + Carrying Case",
		"Latest mixed-reality headset. Includes 4 extra games (Asgard's Wrath 2, Resident Evil 4, Superhot, Beat Saber).",
		"https://images.unsplash.com/photo-1617802690992-15d93263d3a9?auto=format&fit=crop&w=800&q=80",
		400, 650, 25,
	},

	// Watches
	{
		"Rolex Submariner Date 41mm – 126610LV",
		"2022 Rolex Submariner \"Kermit\" in stainless steel. Full set with box, papers, hang tag. Purchased at authorized dealer.",
		"https://images.unsplash.com/photo-1523275335684-37898b6baf30?auto=format&fit=crop&w=800&q=80",
		12000, 18000, 200,
	},
	{
		"Omega Seamaster Diver 300M 42mm",
		"Reference 210.30.42.20.03.001. Blue dial on rubber strap. Full set, 2021 model. James Bond edition.",
		"https://images.unsplash.com/photo-1533139502658-0198f920d8e8?auto=format&fit=crop&w=800&q=80",
		3500, 5500, 100,
	},
	{
		"TAG Heuer Carrera Chronograph 39mm",
		"Ref. CBN2A1B.BA0643. Black dial with ceramic bezel. Full set, lightly worn. Stunning condition.",
		"https://images.unsplash.com/photo-1547996160-81dfa63595aa?auto=format&fit=crop&w=800&q=80",
		3000, 4800, 100,
	},
	{
		"IWC Pilot's Watch Mark XX 40mm",
		"Ref. IW328201. Full stainless steel with black dial. Complete set with box and papers. Worn under 10 times.",
		"https://images.unsplash.com/photo-1618367588411-d9a90fefa881?auto=format&fit=crop&w=800&q=80",
		4000, 6000, 100,
	},
	{
		"Breitling Navitimer B01 Chronograph 46mm",
		"Aviation chronometer with COSC certified movement. Full black dial, blue subdials. Complete set from 2022.",
		"https://images.unsplash.com/photo-1632415786787-abffe7bbf44e?auto=format&fit=crop&w=800&q=80",
		7000, 10000, 200,
	},

	// Art & Collectibles
	{
		"Original Oil Painting – \"City at Dusk\" 36\"×24\"",
		"Large-format original painting on canvas by contemporary artist Marcus Chen. Certificate of authenticity included. Unframed.",
		"https://images.unsplash.com/photo-1541961017774-22349e4a1262?auto=format&fit=crop&w=800&q=80",
		400, 900, 25,
	},
	{
		"Vintage Vinyl Record Collection – 250 LPs",
		"Carefully curated collection spanning jazz, soul, funk, and classic rock. 1960s-1980s. Condition rated VG+ to NM. Full list available.",
		"https://images.unsplash.com/photo-1532649842991-adf7fd68ebe1?auto=format&fit=crop&w=800&q=80",
		600, 1200, 25,
	},
	{
		"LEGO Star Wars Millennium Falcon 75192",
		"Ultimate Collector Series. 7,541 pieces. Sealed in original box. Never opened. Purchased 2019.",
		"https://images.unsplash.com/photo-1585366119957-e9730b6d0f60?auto=format&fit=crop&w=800&q=80",
		700, 1100, 25,
	},
	{
		"Vintage Royal Typewriter – Fully Serviced",
		"1955 Royal Quiet De Luxe in excellent working condition. Professionally cleaned and serviced. All keys function perfectly.",
		"https://images.unsplash.com/photo-1558618666-fcd25c85cd64?auto=format&fit=crop&w=800&q=80",
		300, 600, 25,
	},
	{
		"First Edition \"The Great Gatsby\" – F. Scott Fitzgerald",
		"1925 first edition in original dust jacket. Minor foxing to pages but structurally sound. A true collector's piece.",
		"https://images.unsplash.com/photo-1512820790803-83ca734da794?auto=format&fit=crop&w=800&q=80",
		2000, 4000, 100,
	},
	{
		"Vintage Movie Poster Collection – 1960s Hollywood",
		"Set of 12 original one-sheets from 1960s films including Bullitt, The Graduate, and Bonnie & Clyde. Professionally framed.",
		"https://images.unsplash.com/photo-1440404653325-ab127d49abc1?auto=format&fit=crop&w=800&q=80",
		500, 900, 25,
	},
	{
		"Bronze Sculpture – Abstract Figure 18\"",
		"Hand-cast limited edition (3/12) bronze by sculptor Rina Mori. Signed and dated 2021. Includes certificate and display base.",
		"https://images.unsplash.com/photo-1518998053901-5348d3961a04?auto=format&fit=crop&w=800&q=80",
		1200, 2400, 50,
	},

	// Sneakers & Fashion
	{
		"Nike Air Jordan 1 Retro High OG \"Chicago\" 2022",
		"Size US 10.5. Deadstock. Never worn. Original box with both lace sets. 100% authentic.",
		"https://images.unsplash.com/photo-1542291026-7eec264c27ff?auto=format&fit=crop&w=800&q=80",
		500, 900, 25,
	},
	{
		"Adidas Yeezy Boost 350 V2 \"Zebra\" DS",
		"Size US 11. Deadstock with receipt from CONFIRMED app. Original box, never tried on.",
		"https://images.unsplash.com/photo-1595950653106-6c9ebd614d3a?auto=format&fit=crop&w=800&q=80",
		350, 650, 25,
	},
	{
		"Nike SB Dunk Low \"What The\" – Size 9",
		"Limited release. Worn twice indoors. Minimal creasing. Includes original box, extras laces, and hang tags.",
		"https://images.unsplash.com/photo-1460353581641-37baddab0fa2?auto=format&fit=crop&w=800&q=80",
		600, 1100, 25,
	},
	{
		"Supreme Box Logo Hoodie FW22 – Size L",
		"Navy colorway from FW22 drop. Worn once, washed cold on gentle cycle. BOGO receipt included.",
		"https://images.unsplash.com/photo-1556821840-3a63f15732ce?auto=format&fit=crop&w=800&q=80",
		350, 700, 25,
	},
	{
		"Louis Vuitton Monogram Pochette Métis",
		"2022 production date. Used lightly, all hardware intact. Comes with LV dust bag, box, and receipt.",
		"https://images.unsplash.com/photo-1548036328-c9fa89d128fa?auto=format&fit=crop&w=800&q=80",
		1500, 2500, 50,
	},
	{
		"Canada Goose Expedition Parka – Black, Size M",
		"Men's size medium. Worn one full winter season. No tears, all zippers and snaps functioning. Fur trim in excellent condition.",
		"https://images.unsplash.com/photo-1539533018447-63fcce2678e3?auto=format&fit=crop&w=800&q=80",
		700, 1200, 25,
	},
	{
		"Balenciaga Triple S Sneakers – EU 43",
		"Grey/white/red colorway. Worn 3 times indoors. Comes with box and extra laces.",
		"https://images.unsplash.com/photo-1584735175315-9d5df23be620?auto=format&fit=crop&w=800&q=80",
		500, 900, 25,
	},

	// Jewelry
	{
		"1.5ct Diamond Solitaire Engagement Ring – Platinum",
		"GIA certified, E color, VS1 clarity, Excellent cut. Platinum band. Complete with GIA certificate and original box.",
		"https://images.unsplash.com/photo-1605100804763-247f67b3557e?auto=format&fit=crop&w=800&q=80",
		6000, 10000, 200,
	},
	{
		"18K Gold Cuban Link Chain – 24\" 12mm",
		"Solid 18K yellow gold, 128 grams. Lobster clasp, no repairs or soldering. Stunning everyday chain.",
		"https://images.unsplash.com/photo-1615655406736-b37c4fabf923?auto=format&fit=crop&w=800&q=80",
		3500, 6000, 100,
	},
	{
		"Cartier Love Bracelet – 18K Rose Gold",
		"Size 17. Complete with original screwdriver, box, and papers from 2020. Minor surface scratches only.",
		"https://images.unsplash.com/photo-1573408301185-9521e7d27c58?auto=format&fit=crop&w=800&q=80",
		5500, 9000, 200,
	},
	{
		"Sapphire and Diamond Tennis Bracelet",
		"18K white gold. 7 carats total weight, alternating natural sapphires and round brilliants. Includes appraisal.",
		"https://images.unsplash.com/photo-1535632787350-4e68ef0ac584?auto=format&fit=crop&w=800&q=80",
		2500, 4200, 100,
	},
	{
		"Emerald Cut Emerald Ring – 3ct – 14K Gold",
		"Natural Colombian emerald, 3.1ct, minor inclusions. Set in 14K yellow gold with pavé diamond halo. Certificate included.",
		"https://images.unsplash.com/photo-1602173574767-37ac01994b2a?auto=format&fit=crop&w=800&q=80",
		1800, 3200, 100,
	},

	// Sports & Collectibles
	{
		"Signed Michael Jordan Chicago Bulls Jersey #23",
		"Authentic upper-deck certified autograph. Framed with UV-resistant glass and certificate of authenticity.",
		"https://images.unsplash.com/photo-1546519638492-12a76ee98e15?auto=format&fit=crop&w=800&q=80",
		1500, 3000, 50,
	},
	{
		"Trek Émonda SL 6 Road Bike – 56cm",
		"Carbon frame and fork. Shimano Ultegra drivetrain. Bontrager Aeolus wheelset. Ridden approx 2,000 miles. Excellent condition.",
		"https://images.unsplash.com/photo-1558618666-fcd25c85cd64?auto=format&fit=crop&w=800&q=80",
		2200, 3800, 50,
	},
	{
		"Callaway Paradym Driver + 3-Wood Set",
		"2023 model. 9° driver with stiff Project X HZRDUS shaft. Lightly used, no bag rash. Headcovers included.",
		"https://images.unsplash.com/photo-1535131432816-265daa9f77e0?auto=format&fit=crop&w=800&q=80",
		450, 800, 25,
	},
	{
		"Vintage Baseball Card Collection – PSA Graded",
		"22 PSA-graded cards including several pre-war era. Lowest grade PSA 5. Full list and scans available on request.",
		"https://images.unsplash.com/photo-1574629810360-7efbbe195018?auto=format&fit=crop&w=800&q=80",
		800, 1600, 50,
	},
	{
		"Wilson Pro Staff RF97 Autograph – Roger Federer",
		"Limited edition racket personally signed by Roger Federer. Comes with Wilson COA and display stand.",
		"https://images.unsplash.com/photo-1554068865-24cecd4e34b8?auto=format&fit=crop&w=800&q=80",
		600, 1200, 25,
	},

	// Luxury & Lifestyle
	{
		"Montblanc Meisterstück 149 Fountain Pen – 18K Gold",
		"Legendary writer's tool. Black with gold trim. Medium 18K nib. Comes with Montblanc converter, ink, and box.",
		"https://images.unsplash.com/photo-1586339949916-3e773aef9640?auto=format&fit=crop&w=800&q=80",
		500, 900, 25,
	},
	{
		"Hermès Cashmere-Silk Scarf – Les Chevaux",
		"90cm × 90cm. Ivory with equestrian motif. Unworn with original box and ribbon. 70% cashmere, 30% silk.",
		"https://images.unsplash.com/photo-1578632767115-351597cf2477?auto=format&fit=crop&w=800&q=80",
		600, 1100, 25,
	},
	{
		"Macallan 25 Year Sherry Oak Single Malt",
		"Sealed bottle, 750ml. Purchased from authorized retailer in 2021. Includes original tube and certificate.",
		"https://images.unsplash.com/photo-1569529465841-dfecdab7503b?auto=format&fit=crop&w=800&q=80",
		1500, 2800, 50,
	},
	{
		"Antique Marble Chess Set – Italian Craftsmanship",
		"Hand-carved Carrara marble set. Board 20\"×20\", king height 4\". Late 19th century. Velvet storage box included.",
		"https://images.unsplash.com/photo-1528819622765-d6bcf132f793?auto=format&fit=crop&w=800&q=80",
		400, 800, 25,
	},
	{
		"Baccarat Crystal Harcourt 1841 Decanter Set",
		"Full set: one decanter + 6 whisky tumblers. Never used, still in original packaging. Discontinued colorway.",
		"https://images.unsplash.com/photo-1474797853765-7f581e56c7d3?auto=format&fit=crop&w=800&q=80",
		800, 1500, 50,
	},

	// More Electronics
	{
		"Leica M11 Rangefinder Camera – Black Paint",
		"60MP BSI CMOS sensor, all-metal body. Used on one photo trip. Full set with box, strap, and Leica warranty card.",
		"https://images.unsplash.com/photo-1452780212461-7786a5944041?auto=format&fit=crop&w=800&q=80",
		7000, 11000, 200,
	},
	{
		"Bose Lifestyle 650 Home Entertainment System",
		"5.1 surround sound system with Acoustimass module. Jewel Cube speakers. Used 2 years in living room.",
		"https://images.unsplash.com/photo-1545454675-3531b543be5d?auto=format&fit=crop&w=800&q=80",
		1200, 2200, 50,
	},
	{
		"Nintendo Switch OLED – Zelda Edition + 12 Games",
		"Limited edition Zelda OLED Switch. Includes 12 physical games: BotW, TotK, Metroid, Mario Kart 8, and more.",
		"https://images.unsplash.com/photo-1587131782738-de30ea91a542?auto=format&fit=crop&w=800&q=80",
		500, 850, 25,
	},
	{
		"Wacom Cintiq Pro 27\" Drawing Tablet Display",
		"4K OLED display, 8192 pressure levels. Used for one year in professional studio. Includes Pro Pen 3 and ExpressKey Remote.",
		"https://images.unsplash.com/photo-1558655146-9f40138edfeb?auto=format&fit=crop&w=800&q=80",
		1800, 2800, 50,
	},
	{
		"Herman Miller Embody Gaming Chair – Black",
		"Glacier colorway. Used 2 years, no structural damage. Upholstery in very good condition. Self-assembly manual included.",
		"https://images.unsplash.com/photo-1593642632559-0c6d3fc62b89?auto=format&fit=crop&w=800&q=80",
		900, 1600, 50,
	},

	// More Fashion
	{
		"Air Jordan 4 Retro \"University Blue\" – Size 10",
		"2021 release. Worn twice for photos. Stored in original box with extra laces and hang tags.",
		"https://images.unsplash.com/photo-1491553895911-0055eca6402d?auto=format&fit=crop&w=800&q=80",
		280, 550, 10,
	},
	{
		"Patagonia Down Sweater Vest – Men's XL",
		"800-fill-power down. Classic navy with orange chest logo. Worn one season. No repairs, all zippers working.",
		"https://images.unsplash.com/photo-1478720568477-152d9b164e26?auto=format&fit=crop&w=800&q=80",
		160, 280, 10,
	},
	{
		"New Balance 992 \"Grey\" – Made in USA – Size 11",
		"2023 release. Deadstock. Original box, both lace sets included. One of the cleanest silhouettes NB makes.",
		"https://images.unsplash.com/photo-1539185441755-769473a23570?auto=format&fit=crop&w=800&q=80",
		250, 450, 10,
	},
	{
		"Arc'teryx Beta AR Gore-Tex Jacket – Men's L",
		"Hard-shell touring jacket. Used one ski season. Gore-Tex membrane fully intact. Powder skirt in excellent condition.",
		"https://images.unsplash.com/photo-1467043198406-dc953a3defa0?auto=format&fit=crop&w=800&q=80",
		450, 750, 25,
	},
	{
		"Loewe Puzzle Bag – Small – Tan Calfskin",
		"Iconic Loewe geometric bag. Used 6 times, no wear on corners. Comes with original dustbag, box, and care card.",
		"https://images.unsplash.com/photo-1566150905458-1bf1fc113f0d?auto=format&fit=crop&w=800&q=80",
		1800, 3200, 100,
	},

	// More Watches
	{
		"Seiko Prospex \"Turtle\" SPB317 – Blue Dial",
		"200m diver, Hi-Beat movement. Full set with box and papers from 2023. Condition: near mint.",
		"https://images.unsplash.com/photo-1619134778706-7015533a6150?auto=format&fit=crop&w=800&q=80",
		450, 750, 25,
	},
	{
		"Grand Seiko SBGA211 \"Snowflake\" – Spring Drive",
		"Iconic textured dial, Spring Drive movement. Full set with original bracelet and rubber strap. 2021 purchase.",
		"https://images.unsplash.com/photo-1622434641406-a158123450f9?auto=format&fit=crop&w=800&q=80",
		5000, 8500, 200,
	},

	// Instruments
	{
		"Gibson Les Paul Standard '60s – Iced Tea Burst",
		"2022 model, barely played. Beautiful figured top. Comes with original case, strap, and hang tags.",
		"https://images.unsplash.com/photo-1510915361894-db8b60106cb1?auto=format&fit=crop&w=800&q=80",
		2200, 3600, 50,
	},
	{
		"Roland FP-90X Digital Piano – Black",
		"Flagship portable piano. Triple sensor key action, 700 sounds. Includes stand, bench, damper pedal, and bag.",
		"https://images.unsplash.com/photo-1520523839897-bd0b52f945a0?auto=format&fit=crop&w=800&q=80",
		1600, 2600, 50,
	},
	{
		"Vintage 1972 Fender Stratocaster – Olympic White",
		"All original hardware except tuning machines (Kluson style replacements). Original case. Plays and sounds incredible.",
		"https://images.unsplash.com/photo-1525201548942-d8732f6617a0?auto=format&fit=crop&w=800&q=80",
		7000, 13000, 200,
	},

	// Home & Living
	{
		"Braun Type 3 Coffee Maker – Vintage 1980s",
		"Iconic Dieter Rams design. Fully refurbished and tested. A museum piece that still brews perfect coffee.",
		"https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?auto=format&fit=crop&w=800&q=80",
		200, 450, 10,
	},
	{
		"Vitamix A3500 Ascent Blender – Brushed Stainless",
		"Used twice. Includes multiple containers, tamper, cookbook. Essentially new condition with full warranty.",
		"https://images.unsplash.com/photo-1570197788417-0e82375c9371?auto=format&fit=crop&w=800&q=80",
		400, 700, 25,
	},
	{
		"Sonos Era 300 Spatial Audio Speaker – White",
		"Dolby Atmos enabled. Purchased 5 months ago. All cables and accessories included.",
		"https://images.unsplash.com/photo-1608043152269-423dbba4e7e1?auto=format&fit=crop&w=800&q=80",
		300, 550, 25,
	},

	// Photography
	{
		"Canon EOS R5 Body – Shutter Count: 3,200",
		"45MP full-frame mirrorless. Low shutter count from professional wedding photographer. Complete set with two batteries.",
		"https://images.unsplash.com/photo-1516035069371-29a1b244cc32?auto=format&fit=crop&w=800&q=80",
		2800, 4200, 50,
	},
	{
		"Hasselblad 503CW Medium Format Kit",
		"Body + 80mm f/2.8 CF Planar T*, A12 film back, prism finder, and original case. All in excellent working order.",
		"https://images.unsplash.com/photo-1468436385273-8abca6dfd8d3?auto=format&fit=crop&w=800&q=80",
		2500, 4500, 100,
	},

	// Bags & Travel
	{
		"Rimowa Original Cabin Plus Suitcase – Silver",
		"Aluminium carry-on. Used on 4 trips, no deep scratches. Includes TSA combination lock and all original keys.",
		"https://images.unsplash.com/photo-1553062407-98eeb64c6a62?auto=format&fit=crop&w=800&q=80",
		600, 1100, 25,
	},
	{
		"Chrome Industries Bravo 3.0 Messenger Bag",
		"Mil-spec buckle, 22L main compartment, waterproof liner. Lightly used commuter bag in excellent condition.",
		"https://images.unsplash.com/photo-1553062407-98eeb64c6a62?auto=format&fit=crop&w=800&q=80",
		120, 220, 10,
	},

	// More Art
	{
		"Large Format Print – \"Tokyo Nocturne\" by K. Tanaka",
		"120cm × 80cm archival pigment print on Hahnemühle paper. Edition 4/10. Signed and numbered, certificate included.",
		"https://images.unsplash.com/photo-1540959733332-eab4deabeeaf?auto=format&fit=crop&w=800&q=80",
		600, 1200, 25,
	},
	{
		"Aerial Photography Triptych – Iceland Highlands",
		"Three 60×40cm prints, museum-mounted and ready to hang. Limited to 15 sets worldwide. Signed by photographer.",
		"https://images.unsplash.com/photo-1476474297799-756bba63eefc?auto=format&fit=crop&w=800&q=80",
		700, 1400, 25,
	},

	// Outdoor & Sports
	{
		"Patagonia Nano-Air Light Hybrid Jacket – Women's S",
		"Technical insulated jacket. Trail-worn twice. All seams intact, no abrasion. Ultralight packable down alternative.",
		"https://images.unsplash.com/photo-1501286353178-1ec881214838?auto=format&fit=crop&w=800&q=80",
		180, 320, 10,
	},
	{
		"Osprey Atmos AG 65 Backpacking Pack – Size M",
		"Anti-gravity suspension system. Used on two 4-day trips. Hip belt pockets, sleeping bag compartment. Near new.",
		"https://images.unsplash.com/photo-1478720568477-152d9b164e26?auto=format&fit=crop&w=800&q=80",
		280, 480, 10,
	},
	{
		"Salomon Sense Ride 5 Trail Shoes – Men's 11",
		"30 miles on trails. Excellent grip remaining. No hot spots or repairs. Fresh laces included.",
		"https://images.unsplash.com/photo-1542291026-7eec264c27ff?auto=format&fit=crop&w=800&q=80",
		80, 160, 5,
	},
	{
		"Metolius Camp 5 Home Climbing Wall Kit",
		"Complete DIY board kit for garage wall installation. Includes 100 T-nuts, hardware pack, and instructions. Never assembled.",
		"https://images.unsplash.com/photo-1504280390367-361c6d9f38f4?auto=format&fit=crop&w=800&q=80",
		500, 900, 25,
	},

	// More Jewelry
	{
		"Van Cleef & Arpels Vintage Alhambra Necklace – 10 Motifs",
		"Turquoise on yellow gold. Length 68cm. Full set with pouch, box, and VCA certificate. 2019 purchase.",
		"https://images.unsplash.com/photo-1535632787350-4e68ef0ac584?auto=format&fit=crop&w=800&q=80",
		8000, 14000, 200,
	},
	{
		"Pomellato Nudo Pink Quartz Ring – 18K Rose Gold",
		"Size 53 (US 6.5). Barely worn. Complete with Pomellato box and authenticity card.",
		"https://images.unsplash.com/photo-1602173574767-37ac01994b2a?auto=format&fit=crop&w=800&q=80",
		1200, 2200, 50,
	},

	// Books & Manuscripts
	{
		"Complete Tolkien Deluxe Edition Box Set – Signed Illustrator",
		"The Lord of the Rings + The Hobbit deluxe illustrated editions. All signed by Alan Lee. Mint condition.",
		"https://images.unsplash.com/photo-1524995997946-a1c2e315a42f?auto=format&fit=crop&w=800&q=80",
		800, 1600, 50,
	},
	{
		"Folio Society Special Edition – \"Dune\" Frank Herbert",
		"Limited edition, numbered 0423/1500. Illustrated by Sam Weber. Sealed in original slipcase.",
		"https://images.unsplash.com/photo-1476275466078-4007374efbbe?auto=format&fit=crop&w=800&q=80",
		300, 600, 25,
	},

	// More Collectibles
	{
		"Funko Pop! Star Wars Grail Collection – 22 Pieces",
		"All grails including Gold Boba Fett, Darth Maul, Chrome Stormtrooper. Most still sealed. Full list available.",
		"https://images.unsplash.com/photo-1585366119957-e9730b6d0f60?auto=format&fit=crop&w=800&q=80",
		500, 950, 25,
	},
	{
		"Vintage Coca-Cola Advertising Sign Collection",
		"6 original enamel tin signs from 1940s-1960s. All with original patina, no reproductions. Great condition for age.",
		"https://images.unsplash.com/photo-1535025639604-9a1f0c724f74?auto=format&fit=crop&w=800&q=80",
		350, 700, 25,
	},
	{
		"Hot Wheels Redline Collection – 47 Cars",
		"All original redline era (1968-1977). Includes several rare castings. Condition VG to EX. Store pickup preferred.",
		"https://images.unsplash.com/photo-1594787318286-3d835c1d207f?auto=format&fit=crop&w=800&q=80",
		600, 1300, 50,
	},
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	// Load .env from backend root (two dirs up from cmd/seed/)
	envPath := "../../.env"
	if _, err := os.Stat(envPath); err == nil {
		_ = godotenv.Load(envPath)
	} else {
		_ = godotenv.Load(".env")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "harsh123"),
		getEnv("DB_NAME", "auction"),
		getEnv("DB_PORT", "5432"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	// If --reset flag is passed, wipe existing seed data
	reset := len(os.Args) > 1 && os.Args[1] == "--reset"
	if reset {
		log.Println("--reset: clearing existing bids, auctions, and non-admin users…")
		db.Exec("DELETE FROM bids")
		db.Exec("DELETE FROM auctions")
		db.Exec("DELETE FROM credit_transactions")
		db.Exec("DELETE FROM wallets WHERE user_id IN (SELECT id FROM users WHERE role = 'USER')")
		db.Exec("DELETE FROM users WHERE role = 'USER'")
		log.Println("Cleared.")
	} else {
		// Guard: don't seed twice without --reset
		var userCount int64
		db.Model(&models.User{}).Where("role = ?", "USER").Count(&userCount)
		if userCount >= 20 {
			log.Println("Seed data already present. Pass --reset to wipe and re-seed.")
			return
		}
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	log.Println("Creating 20 users…")
	users := createUsers(db, rng)

	log.Println("Creating 100 auctions…")
	createAuctions(db, rng, users)

	log.Printf("Done! Created %d users and auctions.\n", len(users))
}

// ---------------------------------------------------------------------------
// Create users
// ---------------------------------------------------------------------------

func createUsers(db *gorm.DB, rng *rand.Rand) []models.User {
	_ = rng
	var created []models.User

	hash, err := bcrypt.GenerateFromPassword([]byte("Password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("bcrypt error:", err)
	}

	for _, u := range userNames {
		user := models.User{
			Username:     u.username,
			Email:        u.username + "@example.com",
			PasswordHash: string(hash),
			Role:         "USER",
			IsActive:     true,
		}
		if err := db.Create(&user).Error; err != nil {
			log.Printf("skipping user %s: %v\n", u.username, err)
			continue
		}
		wallet := models.Wallet{
			UserID:  user.ID,
			Balance: 10000,
		}
		if err := db.Create(&wallet).Error; err != nil {
			log.Printf("wallet error for %s: %v\n", u.username, err)
			continue
		}
		created = append(created, user)
		log.Printf("  created user: %s (id=%d)\n", user.Username, user.ID)
	}
	return created
}

// ---------------------------------------------------------------------------
// Create auctions + bids
// ---------------------------------------------------------------------------

func createAuctions(db *gorm.DB, rng *rand.Rand, users []models.User) {
	// Find admin user to be the auction creator
	var admin models.User
	if err := db.Where("role = ?", "ADMIN").First(&admin).Error; err != nil {
		log.Fatal("no admin user found – run the server once first to seed the admin")
	}

	now := time.Now()

	// We'll distribute the 100 auction templates across 4 status buckets.
	// Statuses: 60 ACTIVE, 25 ENDED, 10 SCHEDULED, 5 CANCELLED
	templates := auctionTemplates
	total := len(templates)

	// Shuffle so variety is spread across all statuses
	rng.Shuffle(total, func(i, j int) { templates[i], templates[j] = templates[j], templates[i] })

	type statusBucket struct {
		status string
		count  int
	}
	buckets := []statusBucket{
		{"ACTIVE", 60},
		{"ENDED", 25},
		{"SCHEDULED", 10},
		{"CANCELLED", 5},
	}

	idx := 0
	for _, bucket := range buckets {
		for n := 0; n < bucket.count; n++ {
			tmpl := templates[idx%total]
			idx++

			// Randomise price within template range
			price := randomBetween(rng, tmpl.minPrice, tmpl.maxPrice)
			// Round to nearest increment
			price = (price / tmpl.increment) * tmpl.increment
			if price < tmpl.minPrice {
				price = tmpl.minPrice
			}

			var startTime, endTime time.Time

			switch bucket.status {
			case "ACTIVE":
				startTime = now.Add(-time.Duration(rng.Intn(48)+1) * time.Hour)
				endTime = now.Add(time.Duration(rng.Intn(71)+1) * time.Hour)
			case "ENDED":
				startTime = now.Add(-time.Duration(rng.Intn(240)+72) * time.Hour)
				endTime = now.Add(-time.Duration(rng.Intn(60)+2) * time.Hour)
			case "SCHEDULED":
				startTime = now.Add(time.Duration(rng.Intn(96)+24) * time.Hour)
				endTime = startTime.Add(time.Duration(rng.Intn(72)+24) * time.Hour)
			case "CANCELLED":
				startTime = now.Add(-time.Duration(rng.Intn(120)+24) * time.Hour)
				endTime = startTime.Add(time.Duration(rng.Intn(48)+24) * time.Hour)
			}

			auction := models.Auction{
				Title:         tmpl.title,
				Description:   tmpl.description,
				ImageURL:      tmpl.imageURL,
				StartingPrice: tmpl.minPrice,
				BidIncrement:  tmpl.increment,
				Status:        bucket.status,
				StartTime:     startTime,
				EndTime:       endTime,
				CreatedBy:     admin.ID,
			}

			if err := db.Create(&auction).Error; err != nil {
				log.Printf("  error creating auction %q: %v\n", tmpl.title, err)
				continue
			}

			// Simulate bids for ACTIVE and ENDED auctions
			if bucket.status == "ACTIVE" || bucket.status == "ENDED" {
				simulateBids(db, rng, &auction, users, price, tmpl.increment)
			}

			log.Printf("  [%s] %s  →  current bid $%d\n", bucket.status, tmpl.title, auction.CurrentHighestBid)
		}
	}
}

func simulateBids(db *gorm.DB, rng *rand.Rand, auction *models.Auction, users []models.User, targetBid int64, increment int64) {
	if len(users) == 0 {
		return
	}

	// Determine number of bids
	var numBids int
	if auction.Status == "ENDED" {
		numBids = rng.Intn(12) + 3 // 3-14
	} else {
		numBids = rng.Intn(9) // 0-8 (some active auctions have no bids yet)
	}

	if numBids == 0 {
		return
	}

	// Cap target so we don't exceed user wallets
	if targetBid > 8000 {
		targetBid = 8000
	}

	// Build a bid ladder from starting price up to target
	currentBid := auction.StartingPrice
	var lastBidderIdx int = -1

	for i := 0; i < numBids; i++ {
		// Pick a different bidder than the last one
		var bidderIdx int
		for {
			bidderIdx = rng.Intn(len(users))
			if bidderIdx != lastBidderIdx || len(users) == 1 {
				break
			}
		}
		lastBidderIdx = bidderIdx
		bidder := users[bidderIdx]

		bidAmount := currentBid + increment
		if bidAmount > targetBid && i > 0 {
			break
		}
		if bidAmount > 9500 { // safety cap near wallet limit
			break
		}

		// Add jitter so bids aren't always exactly increment apart
		if rng.Float64() < 0.3 {
			bidAmount += increment * int64(rng.Intn(3)+1)
		}

		createdAt := auction.StartTime.Add(
			time.Duration(float64(auction.EndTime.Sub(auction.StartTime)) * float64(i+1) / float64(numBids+1)),
		)

		bid := models.Bid{
			AuctionID: auction.ID,
			UserID:    bidder.ID,
			Amount:    bidAmount,
			CreatedAt: createdAt,
		}
		if err := db.Create(&bid).Error; err != nil {
			log.Printf("    bid error on auction %d: %v\n", auction.ID, err)
			continue
		}

		currentBid = bidAmount
		auction.CurrentHighestBid = bidAmount
		auction.CurrentHighestBidderID = &bidder.ID
		auction.BidCount++
	}

	// Persist updated auction stats
	db.Model(auction).Updates(map[string]interface{}{
		"current_highest_bid":       auction.CurrentHighestBid,
		"current_highest_bidder_id": auction.CurrentHighestBidderID,
		"bid_count":                 auction.BidCount,
	})

	// For ENDED auctions: deduct the winning bid from winner's wallet
	if auction.Status == "ENDED" && auction.CurrentHighestBidderID != nil {
		db.Model(&models.Wallet{}).
			Where("user_id = ?", *auction.CurrentHighestBidderID).
			UpdateColumn("balance", gorm.Expr("balance - ?", auction.CurrentHighestBid))
	}
}

func randomBetween(rng *rand.Rand, min, max int64) int64 {
	if max <= min {
		return min
	}
	return min + rng.Int63n(max-min)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
