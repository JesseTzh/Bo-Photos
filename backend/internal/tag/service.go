package tag

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, input CreateInput) (Tag, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Tag{}, ErrInvalidName
	}
	now := time.Now()
	item := Tag{ID: newID(), Name: name, Category: name,
		ParentID: input.ParentID, Detail: input.Detail, CreatedAt: now, UpdatedAt: now}
	if input.ParentID != "" {
		parent, err := s.repo.Get(ctx, input.ParentID)
		if err != nil {
			return Tag{}, err
		}
		item.Category = parent.Name
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return Tag{}, err
	}
	return item, nil
}

func (s *Service) Update(ctx context.Context, id string, input UpdateInput) error {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return ErrInvalidName
	}
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	input.Category = input.Name
	if item.ParentID != "" {
		parent, err := s.repo.Get(ctx, item.ParentID)
		if err != nil {
			return err
		}
		input.Category = parent.Name
	}
	return s.repo.Update(ctx, id, input)
}

func (s *Service) Move(ctx context.Context, id, parentID string) error {
	if id == parentID {
		return ErrCycle
	}
	oldAncestors, err := s.repo.AncestorIDs(ctx, id)
	if err != nil {
		return err
	}
	current := parentID
	category := ""
	for current != "" {
		if current == id {
			return ErrCycle
		}
		item, err := s.repo.Get(ctx, current)
		if err != nil {
			return err
		}
		if current == parentID {
			category = item.Name
		}
		current = item.ParentID
	}
	if parentID == "" {
		item, err := s.repo.Get(ctx, id)
		if err != nil {
			return err
		}
		category = item.Name
	}
	if err := s.repo.Move(ctx, id, parentID, category); err != nil {
		return err
	}
	newAncestors, err := s.repo.AncestorIDs(ctx, id)
	if err != nil {
		return err
	}
	keep := map[string]bool{}
	for _, ancestor := range newAncestors {
		keep[ancestor] = true
	}
	obsolete := make([]string, 0)
	for _, ancestor := range oldAncestors {
		if ancestor != id && !keep[ancestor] {
			obsolete = append(obsolete, ancestor)
		}
	}
	if err := s.repo.RemoveAssociationsFromAssetsWithTag(ctx, id, obsolete); err != nil {
		return err
	}
	return s.repo.ResyncAllAssetAncestors(ctx)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	hasChildren, err := s.repo.HasChildren(ctx, id)
	if err != nil {
		return err
	}
	if hasChildren {
		return ErrHasChildren
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) AssignAsset(ctx context.Context, assetID string, ids []string) error {
	seen := map[string]bool{}
	expanded := []string{}
	for _, id := range ids {
		ancestors, err := s.repo.AncestorIDs(ctx, id)
		if err != nil {
			return err
		}
		for _, ancestor := range ancestors {
			if !seen[ancestor] {
				seen[ancestor] = true
				expanded = append(expanded, ancestor)
			}
		}
	}
	return s.repo.ReplaceAssetTags(ctx, assetID, expanded)
}

func Tree(items []Tag) []Tag {
	children := map[string][]Tag{}
	for _, item := range items {
		children[item.ParentID] = append(children[item.ParentID], item)
	}
	var build func(string) []Tag
	build = func(parent string) []Tag {
		nodes := children[parent]
		for i := range nodes {
			nodes[i].Children = build(nodes[i].ID)
		}
		return nodes
	}
	return build("")
}

func newID() string {
	value := make([]byte, 16)
	_, _ = rand.Read(value)
	return hex.EncodeToString(value)
}
