## GitHubAPIScraper
A Golang GitHub API HTTP client to scrape data for a specific user.

The program pretty-prints report with collected data, and generates JSON file with all results.

## Instructions
- The program reads a text file given as a command-line argument;
- The text file contains one of several GitHub usernames;
- If the document contains several usernames, each username must be written on a separate line;
- Run program: `go run main.go <textfile_name>` (example with text file provided in repo: `go run main.go users.txt`).

## Data scraped and statistics report

- The program prints a statistics report as a table containing 6 columns:
1) Username;
2) Number of user repositories;
3) Distribution of programming languages: percentage ratios in the same column. Sum should be 100%, only top 5 should be listed + "other" category for all other languages; 
4) Number of followers;
5) Number of forks for all repositories;
6) Distribution of activity by year: percentage ratios in the same column. Sum should be 100%, only last 5 years should be listed + "other" category for all years before;

At the end of the process a JSON file is generated with report data.
