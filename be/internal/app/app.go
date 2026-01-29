package app

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/http/router"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

type App struct {
	Router http.Handler

	// repositories
	userRepository             repository.UserRepository
	friendRepository           repository.FriendRepository
	groupRepository            repository.GroupRepository
	expenseRepository          repository.ExpenseRepository
	expenseCategoryRepository  repository.ExpenseCategoryRepository
	expenseCommentRepository   repository.ExpenseCommentRepository
	groupActivityRepository    repository.GroupActivityRepository
	balanceRepository          repository.BalanceRepository
	settlementRepository       repository.SettlementRepository
	recurringExpenseRepository repository.RecurringExpenseRepository
	pendingUserRepository      repository.PendingUserRepository
	groupInvitationRepository  repository.GroupInvitationRepository

	// services
	userService             service.UserService
	friendService           service.FriendService
	friendExpenseService    service.FriendExpenseService
	friendSettlementService service.FriendSettlementService
	groupService            service.GroupService
	expenseService          service.ExpenseService
	expenseCategoryService  service.ExpenseCategoryService
	expenseCommentService   service.ExpenseCommentService
	groupActivityService    service.GroupActivityService
	balanceService          service.BalanceService
	settlementService       service.SettlementService
	recurringExpenseService service.RecurringExpenseService
	groupInvitationService  service.GroupInvitationService
}

func New(pool *pgxpool.Pool, queries *sqlc.Queries) *App {
	app := &App{}

	// initialize repositories
	app.userRepository = repository.NewUserRepository(queries)
	app.friendRepository = repository.NewFriendRepository(queries)
	app.groupRepository = repository.NewGroupRepository(pool, queries)
	app.expenseRepository = repository.NewExpenseRepository(pool, queries)
	app.expenseCategoryRepository = repository.NewExpenseCategoryRepository(pool)
	app.expenseCommentRepository = repository.NewExpenseCommentRepository(queries)
	app.groupActivityRepository = repository.NewGroupActivityRepository(pool)
	app.balanceRepository = repository.NewBalanceRepository(queries)
	app.settlementRepository = repository.NewSettlementRepository(queries)
	app.recurringExpenseRepository = repository.NewRecurringExpenseRepository(pool, queries)
	app.pendingUserRepository = repository.NewPendingUserRepository(pool, queries)
	app.groupInvitationRepository = repository.NewGroupInvitationRepository(pool, queries)

	// initialize services
	app.groupActivityService = service.NewGroupActivityService(app.groupActivityRepository) // Initialize early for dependencies

	app.userService = service.NewUserService(app.userRepository)
	app.friendService = service.NewFriendService(app.friendRepository)
	app.friendExpenseService = service.NewFriendExpenseService(app.expenseRepository, app.friendRepository)
	app.friendSettlementService = service.NewFriendSettlementService(app.settlementRepository, app.friendRepository)
	app.groupService = service.NewGroupService(app.groupRepository, app.groupActivityService)
	app.expenseCategoryService = service.NewExpenseCategoryService(app.expenseCategoryRepository)
	app.expenseService = service.NewExpenseService(app.expenseRepository, app.expenseCategoryRepository, app.groupActivityService)
	app.expenseCommentService = service.NewExpenseCommentService(app.expenseCommentRepository, app.expenseService, app.groupActivityService)
	app.balanceService = service.NewBalanceService(app.balanceRepository)
	app.settlementService = service.NewSettlementService(app.settlementRepository, app.groupActivityService)
	app.recurringExpenseService = service.NewRecurringExpenseService(app.recurringExpenseRepository, app.expenseService)
	app.groupInvitationService = service.NewGroupInvitationService(app.groupInvitationRepository, app.pendingUserRepository, app.groupRepository, app.userRepository, app.userService)

	// initialize router
	app.Router = router.New(
		router.WithUserRoutes(app.userService),
		router.WithFriendRoutes(app.friendService, app.friendExpenseService, app.friendSettlementService),
		router.WithGroupRoutes(app.groupService, app.groupInvitationService),
		router.WithExpenseRoutes(app.expenseService),
		router.WithExpenseCategoryRoutes(app.expenseCategoryService),
		router.WithExpenseCommentRoutes(app.expenseCommentService),
		router.WithGroupActivityRoutes(app.groupActivityService),
		router.WithBalanceRoutes(app.balanceService),
		router.WithSettlementRoutes(app.settlementService),
		router.WithRecurringExpenseRoutes(app.recurringExpenseService),
	)

	// debug - dev only
	if os.Getenv("ENV") != "production" {
		router.PrintRoutes(app.Router.(chi.Router))
	}

	return app
}
