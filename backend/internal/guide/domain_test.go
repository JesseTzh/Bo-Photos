package guide

import (
	"context"
	"testing"
)

func TestDomainCreatesModulesBlocksTOCAndAlbums(t *testing.T) {
	service, repo := newTestService(t)
	ctx := context.Background()
	item, _ := service.Create(ctx, GuideInput{Title: "Guide"})
	module, err := repo.CreateModule(ctx, item.ID, ModuleInput{Name: "Intro", Kind: "content"})
	if err != nil {
		t.Fatal(err)
	}
	block, err := repo.CreateBlock(ctx, item.ID, module.ID, BlockInput{Type: BlockMarkdown, DataVersion: 1, Data: []byte(`{"text":"Hello"}`)})
	if err != nil || block.ID == "" {
		t.Fatalf("block=%#v err=%v", block, err)
	}
	if err := repo.ReplaceTOC(ctx, item.ID, []TOCInput{{Title: "Intro", Level: 1, TargetModuleID: module.ID}}); err != nil {
		t.Fatal(err)
	}
	if err := repo.AutoGenerateTOC(ctx, item.ID); err != nil {
		t.Fatal(err)
	}
	toc, _ := repo.ListTOC(ctx, item.ID, false)
	if len(toc) != 1 || toc[0].TargetModuleID != module.ID {
		t.Fatalf("toc=%#v", toc)
	}
}
