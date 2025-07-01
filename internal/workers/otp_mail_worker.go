package workers

// type OtpEmailJobArgs struct {
// 	UserID uuid.UUID
// 	Type   mailer.EmailType
// }

// func (j OtpEmailJobArgs) Kind() string {
// 	return "otp_email"
// }

// // func RegisterMailWorker(
// // 	dispatcher jobs.Dispatcher,
// // 	userFinder UserFinder,
// // 	otpMailer OtpMail,

// // ) {
// // 	worker := NewOtpEmailWorker(userFinder, otpMailer)
// // 	jobs.RegisterWorker(dispatcher, worker)
// // }

// type otpMailWorker struct {

// }

// func NewOtpEmailWorker(user UserFinder, mail OtpMail) jobs.Worker[OtpEmailJobArgs] {
// 	return &otpMailWorker{
// 		mail: mail,
// 		user: user,
// 	}
// }

// // Work implements jobs.Worker.
// func (w *otpMailWorker) Work(ctx context.Context, job *jobs.Job[OtpEmailJobArgs]) error {
// 	fmt.Println("otp mail")
// 	utils.PrettyPrintJSON(job)
// 	user, err := w.user.FindUser(ctx, &stores.UserFilter{Ids: []uuid.UUID{job.Args.UserID}})
// 	if err != nil {
// 		slog.ErrorContext(
// 			ctx,
// 			"error getting user",
// 			slog.Any("error", err),
// 			slog.String("email", user.Email),
// 			slog.String("emailType", job.Args.Type),
// 			slog.String("userId", user.ID.String()),
// 		)
// 		return err
// 	}
// 	err = w.mail.SendOtpEmail(job.Args.Type, ctx, user, nil)
// 	if err != nil {
// 		slog.ErrorContext(
// 			ctx,
// 			"error sending email",
// 			slog.Any("error", err),
// 			slog.String("email", user.Email),
// 			slog.String("emailType", job.Args.Type),
// 			slog.String("userId", user.ID.String()),
// 		)
// 		return err
// 	}

// 	return nil
// }

// var _ jobs.Worker[OtpEmailJobArgs] = (*otpMailWorker)(nil)
