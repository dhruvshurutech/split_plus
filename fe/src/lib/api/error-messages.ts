const errorMessagesByCode: Record<string, string> = {
  'auth.credentials.invalid': 'Email or password is incorrect.',
  'auth.refresh_token.invalid': 'Your session expired. Please sign in again.',
  'auth.token.expired': 'Your session expired. Please sign in again.',
  'auth.token.invalid': 'Please sign in again to continue.',
  'auth.token.revoked':
    'Your session is no longer valid. Please sign in again.',
  'auth.authorization.missing_header': 'Please sign in to continue.',
  'auth.authorization.invalid_format': 'Please sign in to continue.',
  'auth.authorization.unauthorized': 'Please sign in to continue.',
  'conflict.user.email_already_exists':
    'An account with this email already exists.',
  'resource.user.not_found': 'We could not find your account.',
  'validation.user.email.required': 'Please enter your email.',
  'validation.user.email.email': 'Please enter a valid email address.',
  'validation.user.password.required': 'Please enter your password.',
  'validation.auth.email.required': 'Please enter your email.',
  'validation.auth.email.email': 'Please enter a valid email address.',
  'validation.auth.password.required': 'Please enter your password.',
  'validation.auth.refresh_token.required': 'Refresh token is required.',
  'validation.group.name.required': 'Please enter a group name.',
  'validation.group.name.invalid': 'Please enter a valid group name.',
  'validation.group.group_id.invalid': 'The group link is invalid.',
  'validation.invitation.email.required': 'Please enter an email to invite.',
  'validation.invitation.email.email': 'Please enter a valid email to invite.',
  'validation.invitation.group_id.invalid': 'The group link is invalid.',
  'validation.invitation.token.required': 'Invitation token is missing.',
  'validation.invitation.password.required_existing_account':
    'Enter your password to join with your existing account.',
  'validation.invitation.password.required_new_account':
    'Create a password to join this group.',
  'resource.group.not_found': 'This group no longer exists.',
  'resource.invitation.not_found':
    'This invitation link is invalid or expired.',
  'resource.invitation.expired': 'This invitation has expired.',
  'system.invitation.accept_failed':
    'Could not accept this invitation. Please try again or sign in with the invited email.',
  'permission.group.member_required': 'You are not a member of this group.',
  'permission.group.invitation_denied':
    'You do not have permission to invite members to this group.',
  'permission.invitation.email_mismatch':
    'This invitation belongs to a different email address.',
  'conflict.group.member_already_exists':
    'You are already a member of this group.',
}

const categoryFallbackMessages: Record<string, string> = {
  validation: 'Please check the highlighted fields and try again.',
  auth: 'Please sign in again to continue.',
  conflict: 'This action conflicts with current data.',
  resource: 'The requested item was not found.',
  system: 'Something went wrong. Please try again.',
}

export function getUserFacingErrorMessage(
  code: string | undefined,
  fallbackMessage: string,
) {
  if (!code) return fallbackMessage

  const direct = errorMessagesByCode[code]
  if (direct) return direct

  const category = code.split('.')[0]
  const categoryFallback = categoryFallbackMessages[category]
  if (categoryFallback) return categoryFallback

  return fallbackMessage
}
