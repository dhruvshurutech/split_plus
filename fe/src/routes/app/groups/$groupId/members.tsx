import { useEffect, useMemo, useRef, useState } from 'react'
import { Link, createFileRoute } from '@tanstack/react-router'
import { ArrowLeft, Copy } from 'lucide-react'

import type { GroupMember, UserGroup } from '@/lib/api/groups'
import {
  buildInvitationLink,
  createGroupInvitation,
  listGroupMembers,
  listUserGroups,
} from '@/lib/api/groups'
import { Button, buttonVariants } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/app/groups/$groupId/members')({
  component: GroupMembersPage,
})

function GroupMembersPage() {
  const { groupId } = Route.useParams()
  const [group, setGroup] = useState<UserGroup | null>(null)
  const [members, setMembers] = useState<Array<GroupMember>>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteName, setInviteName] = useState('')
  const [inviteRole, setInviteRole] = useState<'member' | 'admin'>('member')
  const [inviteError, setInviteError] = useState<string | null>(null)
  const [inviteLink, setInviteLink] = useState<string | null>(null)
  const [isInviting, setIsInviting] = useState(false)
  const [copyState, setCopyState] = useState<'idle' | 'copied' | 'failed'>(
    'idle',
  )
  const [notice, setNotice] = useState<{
    type: 'success' | 'error'
    message: string
  } | null>(null)
  const noticeTimeoutRef = useRef<number | null>(null)

  const memberSummary = useMemo(() => {
    let active = 0
    let pending = 0
    for (const member of members) {
      if (member.status === 'active') active += 1
      if (member.status === 'pending') pending += 1
    }
    return { active, pending }
  }, [members])

  function showNotice(type: 'success' | 'error', message: string) {
    setNotice({ type, message })
    if (noticeTimeoutRef.current) {
      window.clearTimeout(noticeTimeoutRef.current)
    }
    noticeTimeoutRef.current = window.setTimeout(() => {
      setNotice(null)
      noticeTimeoutRef.current = null
    }, 1800)
  }

  useEffect(() => {
    let mounted = true
    setIsLoading(true)
    setError(null)

    async function loadData() {
      try {
        const [groupList, groupMembers] = await Promise.all([
          listUserGroups(),
          listGroupMembers(groupId, { force: true }),
        ])
        if (!mounted) return
        setGroup(groupList.find((entry) => entry.id === groupId) || null)
        setMembers(groupMembers)
      } catch (err) {
        if (!mounted) return
        setError(
          err instanceof Error ? err.message : 'Unable to load member list.',
        )
      } finally {
        if (mounted) setIsLoading(false)
      }
    }

    loadData()
    return () => {
      mounted = false
    }
  }, [groupId])

  useEffect(
    () => () => {
      if (noticeTimeoutRef.current) {
        window.clearTimeout(noticeTimeoutRef.current)
      }
    },
    [],
  )

  async function handleInvite(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setInviteError(null)
    setCopyState('idle')
    setIsInviting(true)

    try {
      const optimisticEmail = inviteEmail.trim()
      const optimisticName = inviteName.trim()
      const optimisticRole = inviteRole

      const result = await createGroupInvitation(groupId, {
        email: inviteEmail,
        name: inviteName,
        role: inviteRole,
      })
      setInviteLink(buildInvitationLink(result.token))
      setInviteEmail('')
      setInviteName('')
      setInviteRole('member')
      setMembers((prev) => [
        {
          id: '',
          groupId,
          userId: '',
          invitationToken: result.token,
          role: optimisticRole,
          status: 'pending',
          invitedAt: new Date().toISOString(),
          joinedAt: '',
          user: {
            email: optimisticEmail,
            name: optimisticName,
            avatarUrl: '',
          },
        },
        ...prev.filter(
          (member) =>
            !(
              member.status === 'pending' &&
              member.user.email.toLowerCase() === optimisticEmail.toLowerCase()
            ),
        ),
      ])

      void Promise.resolve(listGroupMembers(groupId, { force: true })).then(
        (updatedMembers) => {
          setMembers(updatedMembers)
        },
      )
    } catch (err) {
      setInviteError(
        err instanceof Error ? err.message : 'Unable to create invite link.',
      )
    } finally {
      setIsInviting(false)
    }
  }

  async function handleCopyInviteLink() {
    if (!inviteLink) return
    try {
      await navigator.clipboard.writeText(inviteLink)
      setCopyState('copied')
      showNotice('success', 'Invitation link copied')
    } catch {
      setCopyState('failed')
      showNotice('error', 'Could not copy link')
    }
  }

  async function handleCopyPendingInviteLink(member: GroupMember) {
    if (!member.invitationToken) return
    const link = buildInvitationLink(member.invitationToken)
    try {
      await navigator.clipboard.writeText(link)
      showNotice('success', 'Invitation link copied')
    } catch {
      showNotice('error', 'Could not copy link')
    }
  }

  return (
    <div className="space-y-4">
      <header className="space-y-1">
        <div className="mb-2 flex items-center gap-2">
          <Link
            to="/app/groups/$groupId"
            params={{ groupId }}
            aria-label="Back to group"
            className={cn(
              buttonVariants({ variant: 'outline', size: 'icon' }),
              'size-8 rounded-full',
            )}
          >
            <ArrowLeft className="size-4" />
          </Link>
          <h1 className="text-xl font-semibold">Members</h1>
        </div>
        <p className="text-sm text-muted-foreground">
          {group?.name || 'Group members'}
        </p>
        <p className="text-xs text-muted-foreground">
          {memberSummary.active} active • {memberSummary.pending} pending
        </p>
      </header>

      {notice ? (
        <div className="pointer-events-none fixed inset-x-0 top-4 z-50 mx-auto flex w-full max-w-md justify-center px-4">
          <div
            role="status"
            aria-live="polite"
            className={`rounded-[var(--radius)] border px-3 py-2 text-xs shadow-sm backdrop-blur ${
              notice.type === 'success'
                ? 'border-border/70 bg-background/95 text-foreground'
                : 'border-destructive/40 bg-destructive/15 text-destructive'
            }`}
          >
            {notice.message}
          </div>
        </div>
      ) : null}

      <Card className="border-border/70 bg-card">
        <CardHeader>
          <CardTitle>Member list</CardTitle>
          <CardDescription>
            Pending members include a copyable invitation link.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-2">
          {isLoading ? (
            <p className="text-xs text-muted-foreground">Loading members...</p>
          ) : null}

          {error ? <p className="text-xs text-destructive">{error}</p> : null}

          {!isLoading && !error && members.length === 0 ? (
            <p className="text-xs text-muted-foreground">
              No members found for this group yet.
            </p>
          ) : null}

          {!isLoading && !error
            ? members.map((member) => (
                <div
                  key={
                    member.id ||
                    `${member.userId}-${member.invitationToken}-${member.status}`
                  }
                  className="flex items-center justify-between border-b border-border/70 pb-2 text-xs last:border-b-0 last:pb-0"
                >
                  <div>
                    <p className="font-medium">
                      {member.user.name || member.user.email}
                    </p>
                    <p className="text-muted-foreground">{member.user.email}</p>
                  </div>
                  <div className="text-right flex gap-4">
                    <div className="flex items-center gap-1">
                      <Badge variant="outline" className="capitalize">
                        {member.role}
                      </Badge>
                      <Badge
                        variant={
                          member.status === 'active' ? 'secondary' : 'ghost'
                        }
                        className="capitalize"
                      >
                        {member.status}
                      </Badge>
                    </div>
                    <div className="mt-0.5 flex items-center justify-end gap-1.5">
                      {member.status === 'pending' && member.invitationToken ? (
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          className="size-7 bg-muted/70 hover:bg-muted"
                          aria-label="Copy invitation link"
                          onClick={() => handleCopyPendingInviteLink(member)}
                        >
                          <Copy className="size-3.5" />
                        </Button>
                      ) : null}
                    </div>
                  </div>
                </div>
              ))
            : null}
        </CardContent>
      </Card>

      <Card className="border-border/70 bg-card">
        <CardHeader>
          <CardTitle>Invite people</CardTitle>
          <CardDescription>
            Generate a shareable invite link with token.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <form className="space-y-3" onSubmit={handleInvite}>
            <Input
              required
              type="email"
              placeholder="friend@example.com"
              value={inviteEmail}
              onChange={(e) => setInviteEmail(e.target.value)}
            />
            <Input
              placeholder="Name (optional)"
              value={inviteName}
              onChange={(e) => setInviteName(e.target.value)}
            />
            <select
              value={inviteRole}
              onChange={(e) =>
                setInviteRole(e.target.value as 'member' | 'admin')
              }
              className="h-[var(--control-height)] w-full rounded-[var(--radius)] border border-input bg-transparent px-3 text-sm"
            >
              <option value="member">member</option>
              <option value="admin">admin</option>
            </select>

            {inviteError ? (
              <p className="text-xs text-destructive">{inviteError}</p>
            ) : null}

            <Button type="submit" disabled={isInviting} className="w-full">
              {isInviting ? 'Creating invite...' : 'Create invite link'}
            </Button>
          </form>

          {inviteLink ? (
            <div className="space-y-2 rounded-[var(--radius)] border border-border/70 p-3">
              <p className="text-xs text-muted-foreground">Share this link</p>
              <p className="break-all text-xs">{inviteLink}</p>
              <Button
                type="button"
                variant="outline"
                className="w-full"
                onClick={handleCopyInviteLink}
              >
                <Copy className="size-3.5" />
                {copyState === 'copied'
                  ? 'Copied'
                  : copyState === 'failed'
                    ? 'Copy failed'
                    : 'Copy link'}
              </Button>
            </div>
          ) : null}
        </CardContent>
      </Card>
    </div>
  )
}
