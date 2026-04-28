-- Add partype to champions for resource-type filtering
ALTER TABLE champions ADD COLUMN partype TEXT;

-- Add requires_mana to augments for mana-gated augment filtering
ALTER TABLE augments ADD COLUMN requires_mana INTEGER DEFAULT 0;
