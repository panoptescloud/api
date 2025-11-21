package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/panoptescloud/api/internal/organisations/domain"
	"github.com/panoptescloud/api/internal/organisations/infra/postgres/db"
)

type OrganisationsRepository struct {
	p *pgxpool.Pool
}

func (u *OrganisationsRepository) ByID(id domain.OrganisationID) (*domain.Organisation, error) {
	queries := db.New(u.p)

	pgOrgID := pgtype.UUID{
		Bytes: [16]byte(id.Bytes()),
		Valid: true,
	}
	dbOrg, err := queries.GetOrganisationByID(context.TODO(), pgOrgID)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	dbMembers, err := queries.GetOrganisationMembers(context.TODO(), pgOrgID)

	if err != nil {
		if err == pgx.ErrNoRows {
			// TODO: better error, this should never happen
			return nil, errors.New("no members for organisation")
		}

		return nil, err
	}

	members := make(domain.Members, len(dbMembers))

	for i, dbM := range dbMembers {
		m, err := domain.HydrateMember(
			dbM.MemberID.String(),
			dbM.Role,
		)

		if err != nil {
			//TODO: better error
			return nil, err
		}

		members[i] = m
	}

	return domain.HydrateOrganisation(
		dbOrg.ID.String(),
		dbOrg.Name,
		members,
	)
}

func (u *OrganisationsRepository) ForMember(id domain.MemberID) ([]*domain.Organisation, error) {
	queries := db.New(u.p)

	pgMemberID := pgtype.UUID{
		Bytes: [16]byte(id.Bytes()),
		Valid: true,
	}

	dbOrgs, err := queries.GetOrganisationsForMember(context.TODO(), pgMemberID)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	orgs := make([]*domain.Organisation, len(dbOrgs))

	for _, dbO := range dbOrgs {
		dbMembers, err := queries.GetOrganisationMembers(context.TODO(), dbO.ID)

		if err != nil {
			if err == pgx.ErrNoRows {
				// TODO: better error, it shouldn't be possible
				return nil, errors.New("no members in organisation")
			}

			return nil, err
		}

		members := make(domain.Members, len(dbMembers))

		for i, dbM := range dbMembers {
			m, err := domain.HydrateMember(
				dbM.MemberID.String(),
				dbM.Role,
			)

			if err != nil {
				//TODO: better error
				return nil, err
			}

			members[i] = m
		}

		o, err := domain.HydrateOrganisation(dbO.ID.String(), dbO.Name, members)
		if err != nil {
			return nil, err
		}

		orgs = append(orgs, o)
	}

	return orgs, nil
}

func (u *OrganisationsRepository) Save(org *domain.Organisation) error {
	tx, err := u.p.BeginTx(context.TODO(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	// Ensure rollback if anything fails
	defer func() {
		if err != nil {
			tx.Rollback(context.TODO())
		}
	}()

	queries := db.New(tx)

	err = queries.UpsertOrganisation(context.TODO(), db.UpsertOrganisationParams{
		ID: pgtype.UUID{
			Bytes: [16]byte(org.ID().Bytes()),
			Valid: true,
		},
		Name: org.Name().String(),
	})

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "unique_name" {
				return domain.ErrNameAlreadyInUse{}
			}
		}

		return err
	}

	for _, m := range org.Members() {
		err = queries.UpsertOrganisationMember(context.TODO(), db.UpsertOrganisationMemberParams{
			MemberID: pgtype.UUID{
				Bytes: [16]byte(m.ID().Bytes()),
				Valid: true,
			},
			OrganisationID: pgtype.UUID{
				Bytes: [16]byte(org.ID().Bytes()),
				Valid: true,
			},
			Role: m.Role().String(),
		})

		if err != nil {
			return err
		}
	}

	return tx.Commit(context.TODO())
}

func NewOrganisationsRepository(p *pgxpool.Pool) *OrganisationsRepository {
	return &OrganisationsRepository{
		p: p,
	}
}
