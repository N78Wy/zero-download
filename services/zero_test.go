package services

import "testing"

func TestGetComicPageInfo(t *testing.T) {
	cli := NewClient()
	zd := &ZeroDownload{
		Cookie:  "",
		OutPath: "download",
		Client:  cli,
		Limit:   12,
	}

	comic := zd.GetComicPageInfo("http://www.zerobyw5090.com/plugin.php?id=jameson_manhua&c=index&a=bofang&kuid=14267")

	if len(comic.Pages) < 7 {
		t.Fatalf("There should be no less than 7 chapters. Pages: %d", len(comic.Pages))
	}

	pages := []Page{
		{Name: "1", Total: 195},
		{Name: "2", Total: 193},
		{Name: "3", Total: 193},
		{Name: "4", Total: 193},
	}
	for i := 0; i < 4; i++ {
		cur := comic.Pages[i]
		if cur.Name != pages[i].Name || cur.Total != pages[i].Total {
			t.Fatalf("pages unequal, current: %s, %d, should be: %s, %d", cur.Name, cur.Total, pages[i].Name, pages[i].Total)
		}
	}
}
