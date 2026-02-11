import { Clock3 } from 'lucide-react'

import { cn } from '@/lib/utils'

export function FeatureNotice({
  title,
  description,
  className,
}: {
  title: string
  description: string
  className?: string
}) {
  return (
    <div
      className={cn(
        'rounded-[var(--radius)] border border-border/70 bg-muted/30 p-3',
        className,
      )}
    >
      <div className="flex items-start gap-2">
        <Clock3 className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
        <div className="space-y-0.5">
          <p className="text-xs font-medium">{title}</p>
          <p className="text-xs text-muted-foreground">{description}</p>
        </div>
      </div>
    </div>
  )
}
