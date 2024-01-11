# torrons

torrons is a single-page application written in Go and HTMX that utilizes a classical ELO system to determine the preferred and better-tasting "torró" based on user input.

## how it works

1. **categories**: there are more than a 100 options, so torrons are divided into categories in order to reduce the required amount of votes per user to produce meaningful results.

2. **pairing**: once a category is chosen, a pair of torrons of such category is presented to the user.

3. **selection**: users must select the preferred one from the pair

4. **ELO system**: a version of the classical ELO system is employed to adjust torró ratings based on user preferences and selections.

5. **iterative process**: the process is repeated with different torró pairs. users can vote as many times as desired, but a progres bar is displayed which will fill up after 20 votes. that's the established minumum amount of votes required per user.
