package constants

const RailwayDocsURL = "https://docs.railway.app"
const RailwayURLDefault = "https://railway.app"

var RAILWAY_URL string = RailwayURLDefault

const VersionDefault = "Piped into LDflags on build. You are probably running Railway CLI from source."

var Version string = VersionDefault

func IsDevVersion() bool {
	return Version == VersionDefault
}
