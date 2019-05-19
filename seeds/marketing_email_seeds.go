package seeds

import "fmt"

// GetMarketingEmailSeedsSQLStatements returns insert statements for sample marketing emails
func GetMarketingEmailSeedsSQLStatements() []string {
	marketingEmailSeeds := []string{"someexistingemail@gmail.com"}

	statements := []string{}
	for _, marketingEmail := range marketingEmailSeeds {
		formattedInsertStatement := fmt.Sprintf("insert into marketing_emails (email, created_at_utc, updated_at_utc) values ('%s', (now() at time zone 'utc'), null);", marketingEmail)
		statements = append(statements, formattedInsertStatement)
	}
	return statements
}
