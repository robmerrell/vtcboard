package updaters

// All data updaters should implement this interface
type Updater interface {
	Update() error
}
