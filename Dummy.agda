module Dummy where

-- Run this in Acme: go build && ./acme-agda -v >log.txt >[2=1]

data ℕ : Set where
  zero : ℕ
  succ : ℕ -> ℕ

_+_ : ℕ → ℕ → ℕ
m + n = {!m!} -- Case

_*_ : ℕ → ℕ → ℕ
zero * n = ? -- Auto
succ m * n = {!m +_ !} -- Refine
