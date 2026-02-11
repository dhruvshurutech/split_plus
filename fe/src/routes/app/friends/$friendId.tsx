import { createFileRoute } from '@tanstack/react-router'

import { FeatureNotice } from '@/components/feature-notice'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export const Route = createFileRoute('/app/friends/$friendId')({
  component: FriendLedgerPage,
})

function FriendLedgerPage() {
  const { friendId } = Route.useParams()

  return (
    <div className="space-y-4">
      <header>
        <p className="text-sm text-muted-foreground">Friend ledger</p>
        <h1 className="mt-1 text-xl font-semibold">{friendId}</h1>
      </header>

      <FeatureNotice
        title="This ledger is coming soon"
        description="Live friend transaction history is not available in alpha yet."
      />

      <Card className="border-border/70 bg-card">
        <CardHeader>
          <CardTitle>Friend ledger</CardTitle>
          <CardDescription>Not available in this release.</CardDescription>
        </CardHeader>
        <CardContent className="text-xs text-muted-foreground">
          Use group expenses for now. Friend-level ledgers will ship in a
          follow-up update.
        </CardContent>
      </Card>
    </div>
  )
}
