package jobs

// // JobStatus defines the state of the job
// // 'pending', 'processing', 'done', 'failed'
// type JobStatus string

// const (
// 	JobStatusPending    JobStatus = "pending"
// 	JobStatusProcessing JobStatus = "processing"
// 	JobStatusDone       JobStatus = "done"
// 	JobStatusFailed     JobStatus = "failed"
// )

// type JobRow struct {
// 	ID          uuid.UUID `db:"id" json:"id"`
// 	Kind        string    `db:"kind" json:"kind"`
// 	UniqueKey   *string   `db:"unique_key" json:"unique_key"`
// 	Payload     []byte    `db:"payload" json:"payload"`
// 	Status      JobStatus `db:"status" json:"status"`
// 	RunAfter    time.Time `db:"run_after" json:"run_after"`
// 	Attempts    int64     `db:"attempts" json:"attempts"`
// 	MaxAttempts int64     `db:"max_attempts" json:"max_attempts"`
// 	LastError   *string   `db:"last_error" json:"last_error"`
// 	CreatedAt   time.Time `db:"created_at" json:"created_at"`
// 	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
// }

// type Job[T JobArgs] struct {
// 	*JobRow
// 	Args T
// }

// type JobArgs interface {
// 	Kind() string
// }

// type Worker[T JobArgs] interface {
// 	Work(ctx context.Context, job *Job[T]) error
// }

// type Dispatcher struct {
// 	handlers map[string]func(context.Context, *JobRow) error
// }

// func NewDispatcher() *Dispatcher {
// 	return &Dispatcher{
// 		handlers: make(map[string]func(context.Context, *JobRow) error),
// 	}
// }

// func RegisterWorker[T JobArgs](d *Dispatcher, worker Worker[T]) {
// 	var zero T
// 	kind := zero.Kind()
// 	if _, ok := d.handlers[kind]; ok {
// 		panic("duplicate worker registered: " + kind)
// 	}
// 	d.handlers[kind] = func(ctx context.Context, row *JobRow) error {
// 		var args T
// 		if err := json.Unmarshal(row.Payload, &args); err != nil {
// 			return fmt.Errorf("decode payload: %w", err)
// 		}
// 		return Execute(ctx, worker, &Job[T]{JobRow: row, Args: args})
// 	}
// }

// func (d *Dispatcher) Dispatch(ctx context.Context, row *JobRow) error {
// 	handler, ok := d.handlers[row.Kind]
// 	if !ok {
// 		return fmt.Errorf("no handler for job kind \"%s\"", row.Kind)
// 	}
// 	return handler(ctx, row)
// }

// func Execute[T JobArgs](ctx context.Context, worker Worker[T], job *Job[T]) (err error) {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			log.Printf("panic in job kind=%s: %v\n%s", job.Kind(), r, debug.Stack())
// 			err = fmt.Errorf("panic: %v", r)
// 		}
// 	}()
// 	log.Printf("start job kind=%s id=%s", job.Kind(), job.ID)
// 	err = worker.Work(ctx, job)
// 	if err != nil {
// 		log.Printf("fail job kind=%s id=%s err=%v", job.Kind(), job.ID, err)
// 	} else {
// 		log.Printf("done job kind=%s id=%s", job.Kind(), job.ID)
// 	}
// 	return err
// }

// type JobStore struct {
// 	DB *pgxpool.Pool
// }

// func (s *JobStore) ClaimPendingJobs(ctx context.Context, limit int) ([]JobRow, error) {
// 	rows, err := s.DB.Query(ctx, `
// 		UPDATE jobs SET status='processing', updated_at=clock_timestamp(), attempts=attempts+1
// 		WHERE id IN (
// 			SELECT id FROM jobs
// 			WHERE status = 'pending' AND run_after <= clock_timestamp() AND attempts < max_attempts
// 			ORDER BY run_after
// 			LIMIT $1 FOR UPDATE SKIP LOCKED
// 		)
// 		RETURNING id, kind, unique_key, payload, status, run_after, attempts, max_attempts, last_error, created_at, updated_at
// 	`, limit)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var jobs []JobRow
// 	for rows.Next() {
// 		var row JobRow
// 		if err := rows.Scan(
// 			&row.ID, &row.Kind, &row.UniqueKey, &row.Payload, &row.Status, &row.RunAfter,
// 			&row.Attempts, &row.MaxAttempts, &row.LastError, &row.CreatedAt, &row.UpdatedAt,
// 		); err != nil {
// 			return nil, err
// 		}
// 		jobs = append(jobs, row)
// 	}

// 	return jobs, rows.Err()
// }

// func (s *JobStore) MarkDone(ctx context.Context, id uuid.UUID) error {
// 	_, err := s.DB.Exec(ctx, `
// 		UPDATE jobs SET status='done', updated_at=clock_timestamp() WHERE id=$1
// 	`, id)
// 	return err
// }

// func (s *JobStore) MarkFailed(ctx context.Context, id uuid.UUID, reason string) error {
// 	_, err := s.DB.Exec(ctx, `
// 		UPDATE jobs SET status='failed', last_error=$2, updated_at=clock_timestamp()
// 		WHERE id=$1 AND attempts >= max_attempts
// 	`, id, reason)
// 	return err
// }

// func (s *JobStore) RescheduleJob(ctx context.Context, id uuid.UUID, delay time.Duration) error {
// 	_, err := s.DB.Exec(ctx, `
// 		UPDATE jobs SET run_after = clock_timestamp() + $2, updated_at = clock_timestamp(), status = 'pending'
// 		WHERE id = $1
// 	`, id, delay)
// 	return err
// }

// func (s *JobStore) Enqueue[T JobArgs](ctx context.Context, args T, uniqueKey *string, runAfter time.Time, maxAttempts int) error {
// 	payload, err := json.Marshal(args)
// 	if err != nil {
// 		return fmt.Errorf("marshal args: %w", err)
// 	}
// 	_, err = s.DB.Exec(ctx, `
// 		INSERT INTO jobs (id, kind, unique_key, payload, status, run_after, attempts, max_attempts, created_at, updated_at)
// 		VALUES ($1, $2, $3, $4, 'pending', $5, 0, $6, clock_timestamp(), clock_timestamp())
// 		ON CONFLICT (unique_key)
// 		WHERE status IN ('pending', 'processing')
// 		DO UPDATE SET
// 			payload = EXCLUDED.payload,
// 			run_after = EXCLUDED.run_after,
// 			updated_at = clock_timestamp()
// 	`, uuid.New(), args.Kind(), uniqueKey, payload, runAfter, maxAttempts)
// 	return err
// }

// type Poller struct {
// 	Store      *JobStore
// 	Dispatcher *Dispatcher
// 	Interval   time.Duration
// }

// func (p *Poller) Run(ctx context.Context) error {
// 	ticker := time.NewTicker(p.Interval)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		case <-ticker.C:
// 			if err := p.pollOnce(ctx); err != nil {
// 				log.Printf("poller error: %v", err)
// 			}
// 		}
// 	}
// }

// func (p *Poller) pollOnce(ctx context.Context) error {
// 	jobs, err := p.Store.ClaimPendingJobs(ctx, 10)
// 	if err != nil {
// 		return err
// 	}

// 	for _, row := range jobs {
// 		row := row // capture for closure
// 		go func() {
// 			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
// 			defer cancel()

// 			err := p.Dispatcher.Dispatch(ctx, &row)
// 			if err != nil {
// 				if row.Attempts >= row.MaxAttempts {
// 					_ = p.Store.MarkFailed(ctx, row.ID, err.Error())
// 				} else {
// 					delay := time.Duration(math.Pow(2, float64(row.Attempts))) * time.Second
// 					_ = p.Store.RescheduleJob(ctx, row.ID, delay)
// 				}
// 			} else {
// 				_ = p.Store.MarkDone(ctx, row.ID)
// 			}
// 		}()
// 	}

// 	return nil
// }

// func ServeWithPoller(ctx context.Context, poller *Poller) error {
// 	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
// 	defer cancel()

// 	g, ctx := errgroup.WithContext(ctx)
// 	g.Go(func() error {
// 		return poller.Run(ctx)
// 	})

// 	if err := g.Wait(); err != nil {
// 		log.Printf("graceful shutdown: %v", err)
// 		return err
// 	}
// 	return nil
// }
