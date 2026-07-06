package sharecard

import (
	"sync"
	"testing"
)

// TestConcurrentRendersAreRaceFree exercises the claim in canvas.go's
// package doc: each Render/RenderWrapped/RenderPressKit call builds its
// own canvas (and thus its own font.Face instances), so concurrent HTTP
// requests - which is exactly how these are called in production, see
// internal/http's *_handler.go - never share the mutable state a
// font.Face carries. Run with `go test -race` to make that a real check,
// not just a shape-of-the-code assumption.
func TestConcurrentRendersAreRaceFree(t *testing.T) {
	const goroutines = 16

	var wg sync.WaitGroup
	errs := make(chan error, goroutines*3)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			if _, err := Render(Data{
				HasVotes: true, TotalVotes: i, TopTorroName: "Torró de Xocolata",
				TopTorroRank: 1, RatedTorronCount: i + 1,
			}); err != nil {
				errs <- err
			}
			if _, err := RenderWrapped(WrappedData{
				HasEnoughVotes: true, TotalVotes: i,
				HasBracketVotes: true, BracketRoundsVoted: i, BracketMatchesDecided: i, BracketPicksCorrect: i,
			}); err != nil {
				errs <- err
			}
			if _, err := RenderPressKit(PressKitData{
				HasChampion: true, ChampionName: "Torró Campió", ChampionVotes: i,
			}); err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("concurrent render error: %v", err)
	}
}
