// Copyright 2012 Herbert G. Ficher. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package whois implements a whois client
package whois

import (
	"fmt"
	"net"
	"net/textproto"
	"strings"
)

// A Client represents a client connection to an WHOIS server.
type Client struct {
	// Text is the textproto.Conn used by the Client. It is exported to allow for
	// clients to add extensions.
	Text       *textproto.Conn
	conn       net.Conn
	serverName string
}

// Dial returns a new Client connected to an WHOIS server at addr.
func Dial(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	colon := strings.Index(addr, ":")
	host := addr[:colon]
	return NewClient(conn, host)
}

// NewClient returns a new Client using an existing connection and host as a
// server name to be used when authenticating.
func NewClient(conn net.Conn, host string) (*Client, error) {
	text := textproto.NewConn(conn)
	c := &Client{Text: text, conn: conn, serverName: host}
	return c, nil
}

// Query a domain and return whois results in a slice
func (c *Client) Query(domain string) ([]string, error) {
	id, err := c.Text.Cmd("%s", domain)
	if err != nil {
		return nil, err
	}
	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)
	lines := make([]string, 0)
	for {
		line, err := c.Text.ReadLine()
		if err != nil {
			fmt.Println(err)
			break
		}
		lines = append(lines, line)
	}
	c.Text.Close()
	return lines, nil
}
