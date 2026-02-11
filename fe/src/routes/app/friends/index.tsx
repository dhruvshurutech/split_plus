import { Link, createFileRoute } from '@tanstack/react-router'
import { ArrowRight } from 'lucide-react'

import { FeatureNotice } from '@/components/feature-notice'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export const Route = createFileRoute('/app/friends/')({
  component: FriendsPage,
})

function FriendsPage() {
  return (
    <div className="space-y-4">
      <header>
        <p className="text-sm text-muted-foreground">Friends</p>
        <h1 className="mt-1 text-xl font-semibold">Shared expenses</h1>
      </header>

      <FeatureNotice
        title="Friends module is not live in alpha yet"
        description="Friend ledgers and direct friend settlements are in progress. For now, please use Groups to track and settle expenses."
      />

      <Card className="border-border/70 bg-card">
        <CardHeader>
          <CardTitle>What is coming next</CardTitle>
          <CardDescription>Personal expense sharing with friends.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-2 text-xs text-muted-foreground">
          <p>1. Friend list and invitations</p>
          <p>2. One-to-one expense feed</p>
          <p>3. Direct settle-up flow</p>
          <Link to="/app/groups" className="block pt-1">
            <Button variant="outline" className="w-full">
              Go to groups
              <ArrowRight className="size-4" />
            </Button>
          </Link>
        </CardContent>
      </Card>
    </div>
  )
}
