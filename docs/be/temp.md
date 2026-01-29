# Split+

## Database Schema 

```
users
- id
- email
- password_hash
- name
- avatar_url
- language = 'en'
- created_at
- updated_at
- deleted_at

groups
- id
- name
- description
- metadata {
    type
}
- currency_code
- created_at
- created_by (references users.id)
- updated_at
- updated_by (references users.id)
- deleted_at

group_members
- id
- group_id (references groups.id)
- user_id (references users.id)
- role
- created_at
- updated_at

expenses
- id
- group_id (references groups.id)
- title
- notes
- amount
- currency_code
- date
- created_at
- created_by (references users.id)
- updated_at
- updated_by (references users.id)
- deleted_at

expense_payments
- id
- expense_id (references expenses.id)
- user_id (references users.id)
- amount
- payment_method
- created_at
- updated_at
- deleted_at

expense_split
- id
- expense_id (references expenses.id)
- user_id (references users.id)
- amount_owned
- split_type (equal, percentage, fixed)
- created_at
- updated_at
- deleted_at

settlements
- id
- group_id (references groups.id)
- payer_id (references users.id)
- payee_id (references users.id)
- amount
- currency_code
- paid_at
- notes
- created_at
- updated_at
- deleted_at
```