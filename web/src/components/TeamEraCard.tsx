import { itemsOf, SLOT_LABELS, type Item, type TeamEra } from '../game/types';

interface Props {
  era: TeamEra;
  /** Which item is highlighted but not yet committed. */
  picked: number | null;
  /** Whether an item can legally be taken now. */
  canTake: (kind: number) => boolean;
  onPick: (kind: number) => void;
}

/**
 * One roll: a team-season, and the five things you can take from it.
 *
 * The whole decision is visible at once, which is the point -- a great
 * team-season offers a great car AND a great driver AND a great principal,
 * and seeing them side by side is what makes taking one of them cost
 * something.
 */
export function TeamEraCard({ era, picked, canTake, onPick }: Props) {
  const items = itemsOf(era);
  return (
    <div className="overflow-hidden rounded-lg border border-edge bg-panel" data-testid="team-era">
      <div
        className="flex items-baseline gap-3 border-b border-edge px-4 py-3"
        style={{ borderLeft: `4px solid ${era.livery}` }}
      >
        <span className="font-mono text-2xl tabular-nums" style={{ color: era.livery }}>
          {era.year}
        </span>
        <span className="text-2xl font-semibold tracking-tight" data-testid="roll-team">
          {era.team}
        </span>
      </div>
      <ul>
        {items.map((item) => (
          <ItemRow
            key={item.kind}
            item={item}
            picked={picked === item.kind}
            disabled={!canTake(item.kind)}
            onPick={() => onPick(item.kind)}
          />
        ))}
      </ul>
    </div>
  );
}

function ItemRow({
  item,
  picked,
  disabled,
  onPick,
}: {
  item: Item;
  picked: boolean;
  disabled: boolean;
  onPick: () => void;
}) {
  return (
    <li>
      <button
        type="button"
        onClick={onPick}
        disabled={disabled}
        aria-pressed={picked}
        data-testid={`item-${item.kind}`}
        className={[
          'flex w-full items-center gap-3 border-b border-edge/60 px-4 py-3 text-left transition',
          'last:border-b-0',
          disabled
            ? 'cursor-not-allowed opacity-35'
            : picked
              ? 'bg-accent/15 ring-1 ring-inset ring-accent'
              : 'hover:bg-white/[0.04]',
        ].join(' ')}
      >
        <span className="w-24 shrink-0 text-[11px] uppercase tracking-widest text-muted">
          {SLOT_LABELS[item.slot]}
        </span>
        <span className="min-w-0 flex-1 truncate font-medium">{item.name}</span>
        <span className="hidden gap-3 font-mono text-[11px] tabular-nums text-muted sm:flex">
          {item.stats.map((s) => (
            <span key={s.label}>
              {s.label} {s.value}
            </span>
          ))}
        </span>
        <span className="w-10 shrink-0 text-right font-mono text-lg font-semibold tabular-nums">
          {item.overall}
        </span>
      </button>
    </li>
  );
}
