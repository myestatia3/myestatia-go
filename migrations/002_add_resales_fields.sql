-- Migration: Add Resales integration fields to properties table
-- Created: 2026-02-07

ALTER TABLE properties 
ADD COLUMN IF NOT EXISTS original_price DECIMAL(15,2) DEFAULT NULL COMMENT 'Original price before discounts (for Resales integration)',
ADD COLUMN IF NOT EXISTS plot_m2 DECIMAL(10,2) DEFAULT NULL COMMENT 'Plot/garden/land size in square meters',
ADD COLUMN IF NOT EXISTS terrace_m2 DECIMAL(10,2) DEFAULT NULL COMMENT 'Terrace size in square meters';
