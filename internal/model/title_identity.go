package model

// A struct whose members uniquely identify a title.
type TitleIdentity struct {
	// The Primary SourceFilename for the related title.
	Primary SourceFilename
	// Any SourceFilenames that are not the Primary, but are also considered valid unique keys for the related title
	Aliases []SourceFilename
}

func NewTitleIdentity(
	primary SourceFilename,
	aliases []SourceFilename,
) TitleIdentity {
	return TitleIdentity{Primary: primary, Aliases: aliases}
}

// Determines if other has at least one overlapping name with the current identity
func (id TitleIdentity) Matches(other TitleIdentity) bool {
	for _, name := range id.Names() {
		for _, otherName := range other.Names() {
			if name == otherName {
				return true
			}
		}
	}
	return false
}

// Names returns every name in the identity. Primary plus Aliases
func (id TitleIdentity) Names() []SourceFilename {
	out := make([]SourceFilename, 0, 1+len(id.Aliases))
	out = append(out, id.Primary)
	out = append(out, id.Aliases...)
	return out
}
