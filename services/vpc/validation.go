package vpc

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// The reviewed tag syntax is prose-only in metadata. It remains a handwritten
// extension rather than a service-specific generator branch.
func validateDescribeVpcs(in DescribeVpcsInput) error {
	if in.PageNumber < 0 || in.PageSize < 0 || in.PageSize > 50 || in.OwnerID < 0 || len(in.Tags) > 20 {
		return errors.New("vpc: invalid paging, owner or tag-list parameters")
	}
	if in.VPCID != "" && len(strings.Split(in.VPCID, ",")) > 20 {
		return errors.New("vpc: at most twenty VPC IDs supported")
	}
	for _, tag := range in.Tags {
		if tag.Key == "" || !validTagText(tag.Key) {
			return errors.New("vpc: invalid tag key")
		}
		if tag.Value != nil && !validTagText(*tag.Value) {
			return errors.New("vpc: invalid tag value")
		}
	}
	return nil
}

func validTagText(text string) bool {
	return utf8.ValidString(text) && utf8.RuneCountInString(text) <= 128 && !strings.HasPrefix(text, "aliyun") && !strings.HasPrefix(text, "acs:") && !strings.Contains(text, "http://") && !strings.Contains(text, "https://")
}
