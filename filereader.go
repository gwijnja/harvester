package harvester

type FileReader interface {
	SetNext(FileWriter)
	List() ([]string, error)
	Process(filename string) error
}
