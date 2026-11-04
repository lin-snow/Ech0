// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package oauthserver

import (
	"encoding/json"

	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

// Client is a dynamically registered OAuth client. CIMD clients are resolved
// from their metadata URL and never persisted, so this table only holds
// registrations created through /oauth/register (the deprecated, compatibility
// path).
type Client struct {
	ID                      string `gorm:"type:char(36);primaryKey" json:"-"`
	ClientID                string `gorm:"size:255;not null;uniqueIndex" json:"client_id"`
	ClientName              string `gorm:"size:255" json:"client_name"`
	ClientURI               string `gorm:"size:512" json:"client_uri,omitempty"`
	LogoURI                 string `gorm:"size:512" json:"logo_uri,omitempty"`
	RedirectURIs            string `gorm:"type:text;not null" json:"-"`
	GrantTypes              string `gorm:"type:text" json:"-"`
	ResponseTypes           string `gorm:"type:text" json:"-"`
	TokenEndpointAuthMethod string `gorm:"size:32" json:"token_endpoint_auth_method"`
	ApplicationType         string `gorm:"size:32" json:"application_type,omitempty"`
	CreatedAt               int64  `gorm:"autoCreateTime" json:"client_id_issued_at"`
	UpdatedAt               int64  `gorm:"autoUpdateTime" json:"-"`
}

func (c *Client) TableName() string { return "oauth_clients" }

func (c *Client) BeforeCreate(_ *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuidUtil.NewV7()
	}
	return nil
}

func (c Client) URIs() []string {
	return decodeStrings(c.RedirectURIs)
}

func (c Client) Metadata() ClientMetadata {
	return ClientMetadata{
		ClientID:                c.ClientID,
		ClientName:              c.ClientName,
		ClientURI:               c.ClientURI,
		LogoURI:                 c.LogoURI,
		RedirectURIs:            c.URIs(),
		GrantTypes:              decodeStrings(c.GrantTypes),
		ResponseTypes:           decodeStrings(c.ResponseTypes),
		TokenEndpointAuthMethod: c.TokenEndpointAuthMethod,
		ApplicationType:         c.ApplicationType,
	}
}

// NewClientFromMetadata persists the serialisable fields of a registration.
func NewClientFromMetadata(meta ClientMetadata, clientID string, issuedAt int64) (Client, error) {
	redirects, err := encodeStrings(meta.RedirectURIs)
	if err != nil {
		return Client{}, err
	}
	grants, err := encodeStrings(meta.GrantTypes)
	if err != nil {
		return Client{}, err
	}
	responses, err := encodeStrings(meta.ResponseTypes)
	if err != nil {
		return Client{}, err
	}
	return Client{
		ClientID:                clientID,
		ClientName:              meta.ClientName,
		ClientURI:               meta.ClientURI,
		LogoURI:                 meta.LogoURI,
		RedirectURIs:            redirects,
		GrantTypes:              grants,
		ResponseTypes:           responses,
		TokenEndpointAuthMethod: meta.TokenEndpointAuthMethod,
		ApplicationType:         meta.ApplicationType,
		CreatedAt:               issuedAt,
	}, nil
}

func encodeStrings(values []string) (string, error) {
	if values == nil {
		values = []string{}
	}
	raw, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func decodeStrings(raw string) []string {
	if raw == "" {
		return nil
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil
	}
	return values
}
