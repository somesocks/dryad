package core

type RootRequirementTargetKind string

const (
	RootRequirementTargetKindRoot RootRequirementTargetKind = "root"
	RootRequirementTargetKindEnv  RootRequirementTargetKind = "env"
	RootRequirementTargetKindFile RootRequirementTargetKind = "file"
)

func rootRequirementTargetKind(kind RootRequirementTargetKind) RootRequirementTargetKind {
	if kind == "" {
		return RootRequirementTargetKindRoot
	}
	return kind
}
