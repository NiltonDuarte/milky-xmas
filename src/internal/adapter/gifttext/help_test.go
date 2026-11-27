package gifttext

import (
	"strings"
	"testing"
)

func TestHelpByAccess(t *testing.T) {
	for _, tt := range []struct {
		name         string
		access       HelpAccess
		want, absent []string
	}{
		{"player", HelpAccess{}, []string{"/daily claim", "/look in-your-stocking", "/help me-milky"}, []string{"/give gift", "/grinch summon", "/setup"}},
		{"moderator", HelpAccess{Moderator: true}, []string{"/daily claim", "/give gift", "/grinch summon"}, []string{"/setup"}},
		{"manager", HelpAccess{Manager: true}, []string{"/daily claim", "/setup view", "/setup channel"}, []string{"/give gift", "/grinch summon"}},
		{"both", HelpAccess{Moderator: true, Manager: true}, []string{"/give gift", "/setup channel"}, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Help(tt.access)
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("help is missing %q:\n%s", want, got)
				}
			}
			for _, absent := range tt.absent {
				if strings.Contains(got, absent) {
					t.Errorf("help should not list %q:\n%s", absent, got)
				}
			}
		})
	}
}
