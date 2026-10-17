package catalog

// Images are GitHub's own badge artwork, in the order default, bronze, silver, gold.
// They are loaded from GitHub's CDN (cross-origin-resource-policy: cross-origin), never
// copied. Collected from the fixtures; the weekly canary checks each still loads.
var Images = map[string][]string{
	"pull-shark": {
		"https://github.githubassets.com/assets/pull-shark-default-498c279a747d.png",
		"https://github.githubassets.com/assets/pull-shark-bronze-a37accb528d1.png",
		"https://github.githubassets.com/assets/pull-shark-silver-0643f87ac9fd.png",
		"https://github.githubassets.com/assets/pull-shark-gold-90985540b385.png",
	},
	"starstruck": {
		"https://github.githubassets.com/assets/starstruck-default-b6610abad518.png",
		"https://github.githubassets.com/assets/starstruck-bronze-b71ec4d7766a.png",
		"https://github.githubassets.com/assets/starstruck-silver-275c101223c0.png",
		"https://github.githubassets.com/assets/starstruck-gold-a35e077693cc.png",
	},
	"pair-extraordinaire": {
		"https://github.githubassets.com/assets/pair-extraordinaire-default-579438a20e01.png",
		"https://github.githubassets.com/assets/pair-extraordinaire-bronze-797214f41e7e.png",
		"https://github.githubassets.com/assets/pair-extraordinaire-silver-a0b2b472348c.png",
		"https://github.githubassets.com/assets/pair-extraordinaire-gold-5335a3b8ae7d.png",
	},
	"galaxy-brain": {
		"https://github.githubassets.com/assets/galaxy-brain-default-847262c21056.png",
		"https://github.githubassets.com/assets/galaxy-brain-bronze-da3c4829fd22.png",
		"https://github.githubassets.com/assets/galaxy-brain-silver-2ce8da282a85.png",
		"https://github.githubassets.com/assets/galaxy-brain-gold-9c2c03fd1acf.png",
	},
	"finishing-touch": {
		"https://github.githubassets.com/assets/finishing-touch-default-764b3a996d62.png",
		"https://github.githubassets.com/assets/finishing-touch-bronze-7c49185628f3.png",
		"https://github.githubassets.com/assets/finishing-touch-silver-75ab9a5c22ba.png",
		"https://github.githubassets.com/assets/finishing-touch-gold-e19e1fcbfee2.png",
	},
	"quickdraw":                     {"https://github.githubassets.com/assets/quickdraw-default-39c6aec8ff89.png"},
	"yolo":                          {"https://github.githubassets.com/assets/yolo-default-be0bbff04951.png"},
	"public-sponsor":                {"https://github.githubassets.com/assets/public-sponsor-default-9fa68986b057.png"},
	"arctic-code-vault-contributor": {"https://github.githubassets.com/assets/arctic-code-vault-contributor-default-df8d74122a06.png"},
	"mars-2020-contributor":         {"https://github.githubassets.com/assets/mars-2020-contributor-default-c9cad8495acf.png"},
	"proxima-pioneer":               {"https://github.githubassets.com/assets/proxima-pioneer-default-51cd4e4969d5.png"},
	"proxima-staffshipper":          {"https://github.githubassets.com/assets/proxima-staffshipper-default-84f658288021.png"},
	"proxima-staffuser":             {"https://github.githubassets.com/assets/proxima-staffuser-default-02bf6163ea31.png"},
}

// Image is the art for slug at tier (0 = not earned, shown as the default art).
func Image(slug string, tier int) string {
	art := Images[slug]
	if len(art) == 0 {
		return ""
	}
	i := min(max(tier-1, 0), len(art)-1)
	return art[i]
}
