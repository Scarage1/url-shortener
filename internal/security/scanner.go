package security

// URLScanner checks whether a destination URL is safe to shorten.
type URLScanner interface {
	Check(url string) error
}

type ChainScanner struct {
	scanners []URLScanner
}

func NewChainScanner(scanners ...URLScanner) URLScanner {

	filtered := make([]URLScanner, 0, len(scanners))

	for _, scanner := range scanners {
		if scanner != nil {
			filtered = append(filtered, scanner)
		}
	}

	return &ChainScanner{
		scanners: filtered,
	}
}

func (s *ChainScanner) Check(url string) error {

	for _, scanner := range s.scanners {
		if err := scanner.Check(url); err != nil {
			return err
		}
	}

	return nil
}

type AllowAllScanner struct{}

func (AllowAllScanner) Check(string) error {

	return nil
}
