package main

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9\.]{2,31}$`)

var (
	errUsernameInvalid    = errors.New("invalid username format")
	errUsernameDisallowed = errors.New("disallowed username")
)

func VerifyUsername(username string) error {
	if !usernamePattern.MatchString(username) {
		return errors.Wrapf(errUsernameInvalid, "VerifyUsername: username did not match allowed pattern %q", usernamePattern.String())
	}

	for k, a := range usernameBlacklist {
		for _, e := range a {
			if strings.EqualFold(username, e) {
				return errors.Wrapf(errUsernameDisallowed, "VerifyUsername: disallowed username is in %q group", k)
			}
		}
	}

	return nil
}

// this list is based on https://github.com/marteinn/The-Big-Username-Blacklist
// v1.4.0, provided under the MIT license

var usernameBlacklist = map[string][]string{
	"privileges": []string{
		"abuse", "admin", "administrator", "administration", "autoconfig",
		"broadcasthost", "domain", "editor", "guest", "host", "hostmaster",
		"info", "localdomain", "localhost", "master", "mail", "mail0", "mail1",
		"mail2", "mail3", "mail4", "mail5", "mail6", "mail7", "mail8", "mail9",
		"mailerdaemon", "mailer-daemon", "me", "member", "members", "moderator",
		"nobody", "owner", "postmaster", "poweruser", "private", "public", "root",
		"rootuser", "super", "superuser", "sysadmin", "user", "users", "usenet",
		"webmaster", "www-data",
	},
	"code": []string{
		"400", "401", "403", "404", "405", "406", "407", "408", "409", "410",
		"411", "412", "413", "414", "415", "416", "417", "421", "422", "423",
		"424", "426", "428", "429", "431", "500", "501", "502", "503", "504",
		"505", "506", "507", "508", "509", "510", "511", "aes128-ctr",
		"aes192-ctr", "aes256-ctr", "aes128-gcm", "aes256-gcm", "ajax", "alpha",
		"amp", "asc", "assets", "atom", "beta", "captcha", "cdn", "cgi", "cgi-bin",
		"chacha20-poly1305", "client", "count", "curve25519-sha256", "css",
		"db", "debug", "desc", "dev", "diffie-hellman-group-exchange-sha256",
		"diffie-hellman-group14-sha1", "dns", "dns0", "dns1", "dns2", "dns3",
		"dns4", "ecdh-sha2-nistp256", "ecdh-sha2-nistp384", "ecdh-sha2-nistp521",
		"error", "errors", "exception", "false", "filter", "ftp", "get", "hmac-sha",
		"hmac-sha1", "hmac-sha1-etm", "hmac-sha2-256", "hmac-sha2-512",
		"hmac-sha2-256-etm", "hmac-sha2-512-etm", "http", "httpd", "https",
		"icons", "imap", "img", "is", "isatap", "it", "js", "json", "limit",
		"map", "mis", "mx", "network", "noc", "none", "nil", "ns", "ns0", "ns1",
		"ns2", "ns3", "ns4", "ns5", "ns6", "ns7", "ns8", "ns9", "null", "post",
		"rsa-sha2-512", "rsa-sha2-2", "rss", "script", "source", "sitemap",
		"smtp", "sql", "ssh", "ssh-rsa", "ssl", "ssladmin", "ssladministrator",
		"sslwebmaster", "stage", "staging", "static", "stylesheet", "stylesheets",
		"subdomain", "sudo", "telnet", "test", "true", "umac-64", "umac-128",
		"umac-64-etm", "umac-128-etm", "undefined", "uucp", "var", "void",
		"webmail", "wpad", "www", "www1", "www2", "www3", "www4", "zlib",
	},
	"terms": []string{
		"api", "auth", "authentication", "authorize", "avatar", "banner",
		"banners", "cache", "cas", "config", "cookies", "draft", "email", "head",
		"header", "htpasswd", "mobile", "noreply", "no-reply", "oauth", "oauth2",
		"online", "openid", "passwd", "password", "pop", "pop3", "postfix",
		"redirect", "request", "response", "sdk", "shop", "security", "session",
		"sessions", "system", "tablet", "trial", "username", "yourname",
		"yourusername",
	},
	"financial": []string{
		"advertise", "advertising", "affiliate", "affiliates", "billing",
		"billings", "business", "campaign", "contest", "cart", "checkout",
		"deals", "hosting", "investors", "invoice", "licensing", "marketing",
		"marketplace", "offer", "offers", "order", "orders", "pay", "payment",
		"payments", "partners", "premium", "sales", "store",
	},
	"sections": []string{
		"about", "about-us", "access", "alert", "alerts", "analytics", "app",
		"apps", "account", "accounts", "backup", "board", "bookmark", "bookmarks",
		"blog", "blogs", "calendar", "careers", "category", "categories", "chat",
		"channel", "channels", "comment", "comments", "community", "contact",
		"copyright", "dashboard", "developer", "developers", "docs",
		"documentation", "download", "downloads", "enterprise", "event", "events",
		"example", "explore", "extensions", "family", "faq", "faqs", "features",
		"feed", "feeds", "feedback", "feeds", "file", "files", "fonts", "form",
		"forms", "forum", "forums", "friend", "friends", "follower", "followers",
		"following", "forgot", "forgotpassword", "forgot-password", "group",
		"groups", "guidelines", "guides", "help", "home", "images", "invitations",
		"invite", "invites", "issues", "jobs", "legal", "local", "lost-password",
		"media", "message", "messages", "more", "my", "new", "news", "next",
		"newsletter", "newsletters", "notification", "notifications", "overview",
		"page", "pages", "photo", "photos", "plans", "plugins", "policy",
		"policies", "popular", "portfolio", "press", "pricing", "privacy",
		"privacy-policy", "preferences", "previous", "product", "profile",
		"profiles", "project", "projects", "quota", "refund", "refunds",
		"registration", "replies", "reply", "request-password", "reset-password",
		"return", "returns", "reviews", "rules", "settings", "setup", "services",
		"site", "sites", "stat", "stats", "status", "statistics", "survey", "tag",
		"tags", "team", "terms", "terms-of-use", "testimonials", "theme",
		"themes", "today", "tools", "topic", "topics", "tour", "training",
		"translations", "trending", "widget", "widgets", "you", "video",
		"website", "wiki",
	},
	"actions": []string{
		"add", "buy", "change", "clear", "close", "compare", "compose", "connect",
		"copy", "create", "customize", "delete", "disconnect", "discuss",
		"downvote", "drop", "edit", "exit", "export", "follow", "go", "hide",
		"import", "insert", "invite", "join", "learn", "load", "lock", "login",
		"logout", "modify", "print", "purchase", "put", "reduce", "register",
		"remove", "report", "reset", "review", "save", "search", "select",
		"share", "shift", "signin", "signup", "sort", "subscribe", "support",
		"sync", "translate", "unfollow", "update", "upgrade", "unsubscribe",
		"verify", "view", "vote", "write",
	},
}
