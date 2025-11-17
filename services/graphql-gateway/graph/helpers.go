package graph

import (
	"github.com/quangdang46/NFT-Marketplace/services/graphql-gateway/graph/model"
	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
)

// Helper functions for converting between proto and GraphQL models

func stringPtrToValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func valueToStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func profileFromProto(p *pb.Profile) *model.Profile {
	if p == nil {
		return nil
	}
	return &model.Profile{
		UserID:      p.UserId,
		Username:    valueToStringPtr(p.Username),
		DisplayName: valueToStringPtr(p.DisplayName),
		AvatarURL:   valueToStringPtr(p.AvatarUrl),
		BannerURL:   valueToStringPtr(p.BannerUrl),
		Bio:         valueToStringPtr(p.Bio),
		Locale:      valueToStringPtr(p.Locale),
		Timezone:    valueToStringPtr(p.Timezone),
		SocialsJSON: valueToStringPtr(p.SocialsJson),
		UpdatedAt:   valueToStringPtr(p.UpdatedAt),
	}
}

func walletLinkFromProto(w *pb.WalletLink) *model.WalletLink {
	if w == nil {
		return nil
	}
	return &model.WalletLink{
		ID:         w.Id,
		UserID:     w.UserId,
		AccountID:  w.AccountId,
		Address:    w.Address,
		ChainID:    w.ChainId,
		IsPrimary:  w.IsPrimary,
		VerifiedAt: valueToStringPtr(w.VerifiedAt.String()),
		CreatedAt:  w.CreatedAt.String(),
		UpdatedAt:  w.UpdatedAt.String(),
	}
}
