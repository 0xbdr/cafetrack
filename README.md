Bet. Short + human.


---

☕ cafetrack

I eat at the same café every day and pay later.
Paper logs kept disappearing. So I built this.

cafetrack is a minimal CLI ledger to track what you owe.

No deps. Single binary. Fast.


---

Commands

cafetrack add "tagine" 25.00
cafetrack listunpaid
cafetrack balance
cafetrack pay <id>
cafetrack pay -p <price>
cafetrack log
cafetrack passwd
cafetrack wipe
cafetrack help


---

How it works

Every item gets an ID

Prices stored in cents (no float bugs)

Partial payments use FIFO logic

Balance rolls over automatically

Everything stored in ~/.cafetrack/


Built because paper is unreliable and apps are bloated.

Simple. Fast. Done.
