package app

import (
	"net/http"
	"os"
	"time"

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
	sessionRepository          repository.SessionRepository
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
	themeRepository            repository.ThemeRepository

	// services
	userService             service.UserService
	jwtService              service.JWTService
	authService             service.AuthService
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
	themeService            service.ThemeService
}

func New(pool *pgxpool.Pool, queries *sqlc.Queries, jwtSecret string, accessTokenExpiry, refreshTokenExpiry time.Duration) *App {
	app := &App{}

	// initialize repositories
	app.userRepository = repository.NewUserRepository(queries)
	app.sessionRepository = repository.NewSessionRepository(queries)
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
	app.themeRepository = repository.NewThemeRepository(queries)

	// initialize services
	app.groupActivityService = service.NewGroupActivityService(app.groupActivityRepository) // Initialize early for dependencies

	app.userService = service.NewUserService(app.userRepository)
	app.jwtService = service.NewJWTService(jwtSecret, accessTokenExpiry, refreshTokenExpiry)
	app.authService = service.NewAuthService(app.userService, app.sessionRepository, app.jwtService, accessTokenExpiry, refreshTokenExpiry)
	app.friendService = service.NewFriendService(app.friendRepository)
	app.friendExpenseService = service.NewFriendExpenseService(app.expenseRepository, app.friendRepository)
	app.friendSettlementService = service.NewFriendSettlementService(app.settlementRepository, app.friendRepository)
	app.groupService = service.NewGroupService(app.groupRepository, app.groupInvitationRepository, app.groupActivityService)
	app.expenseCategoryService = service.NewExpenseCategoryService(app.expenseCategoryRepository)
	app.expenseService = service.NewExpenseService(app.expenseRepository, app.expenseCategoryRepository, app.groupActivityService, app.userRepository, app.pendingUserRepository)
	app.expenseCommentService = service.NewExpenseCommentService(app.expenseCommentRepository, app.expenseService, app.groupActivityService)
	app.balanceService = service.NewBalanceService(app.balanceRepository)
	app.settlementService = service.NewSettlementService(app.settlementRepository, app.groupActivityService)
	app.recurringExpenseService = service.NewRecurringExpenseService(app.recurringExpenseRepository, app.expenseService)
	app.groupInvitationService = service.NewGroupInvitationService(app.groupInvitationRepository, app.pendingUserRepository, app.groupRepository, app.userRepository, app.userService)
	app.themeService = service.NewThemeService(app.themeRepository)

	// initialize router
	app.Router = router.New(
		router.WithAuthRoutes(app.authService, app.jwtService, app.sessionRepository),
		router.WithUserRoutes(app.userService, app.themeService, app.jwtService, app.sessionRepository),
		router.WithThemeRoutes(app.themeService, app.jwtService, app.sessionRepository),
		router.WithFriendRoutes(app.friendService, app.friendExpenseService, app.friendSettlementService, app.jwtService, app.sessionRepository),
		router.WithGroupRoutes(app.groupService, app.groupInvitationService, app.jwtService, app.sessionRepository),
		router.WithExpenseRoutes(app.expenseService, app.jwtService, app.sessionRepository),
		router.WithExpenseCategoryRoutes(app.expenseCategoryService, app.jwtService, app.sessionRepository),
		router.WithExpenseCommentRoutes(app.expenseCommentService, app.jwtService, app.sessionRepository),
		router.WithGroupActivityRoutes(app.groupActivityService, app.jwtService, app.sessionRepository),
		router.WithBalanceRoutes(app.balanceService, app.jwtService, app.sessionRepository),
		router.WithSettlementRoutes(app.settlementService, app.jwtService, app.sessionRepository),
		router.WithRecurringExpenseRoutes(app.recurringExpenseService, app.jwtService, app.sessionRepository),
	)

	// debug - dev only
	if os.Getenv("ENV") != "production" {
		router.PrintRoutes(app.Router.(chi.Router))
	}

	return app
}
