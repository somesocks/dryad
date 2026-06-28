package core

type RootRequirementTargetKind string

const (
	RootRequirementTargetKindRoot RootRequirementTargetKind = "root"
	RootRequirementTargetKindEnv  RootRequirementTargetKind = "env"
	RootRequirementTargetKindFile RootRequirementTargetKind = "file"
	RootRequirementTargetKindHTTP RootRequirementTargetKind = "http"
)

func rootRequirementTargetKind(kind RootRequirementTargetKind) RootRequirementTargetKind {
	if kind == "" {
		return RootRequirementTargetKindRoot
	}
	return kind
}
