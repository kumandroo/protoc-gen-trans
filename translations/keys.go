package translations

// KeyGetter is an interface used in protoc-gen-trans autogenerated code
type KeyGetter interface {
	// GetKey takes the current key for a field and new translated text for the field and returns the key
	// to be used for storing the translation.
	GetKey(currentKey, translatedText string) string
}
