module Dummy where

-- Run this in Acme: go build && ./acme-agda

data ℕ : Set where
  zero : ℕ
  succ : ℕ -> ℕ

_+_ : ℕ → ℕ → ℕ
m + n = ?

_*_ : ℕ → ℕ → ℕ
m * n = {! succ !}