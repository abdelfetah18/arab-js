package artspath

type Extension = string

const (
	ExtensionArTs   Extension = ".arts"
	ExtensionDarts  Extension = ".d.arts"
	ExtensionCode   Extension = ".كود"
	ExtensionTa3rif Extension = ".تعريف"
	ExtensionJson   Extension = ".json"
)

func ExtensionIsArTs(ext string) bool {
	return ext == ExtensionArTs || ext == ExtensionDarts || ext == ExtensionCode
}
