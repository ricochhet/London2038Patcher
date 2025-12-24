package patchutil

type Options struct {
	Registry   *LocaleRegistry
	Filter     *LocaleFilter
	IdxOptions *IdxOptions
	Archs      []string
}

type IdxOptions struct {
	Debug bool
	CRC32 bool // PackWithIndex
}
