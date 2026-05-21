-- 004_savings_and_loans.sql
-- Add savings and loan account types with extra metadata columns

ALTER TABLE accounts
  DROP CONSTRAINT IF EXISTS accounts_account_type_check;

ALTER TABLE accounts
  ADD CONSTRAINT accounts_account_type_check
    CHECK (account_type IN ('cash', 'bank_card', 'e_wallet', 'savings', 'loan'));

ALTER TABLE accounts
  ADD COLUMN IF NOT EXISTS interest_rate     NUMERIC(8,4),
  ADD COLUMN IF NOT EXISTS target_amount     NUMERIC(15,4),
  ADD COLUMN IF NOT EXISTS maturity_date     DATE,
  ADD COLUMN IF NOT EXISTS loan_total_amount NUMERIC(15,4);
