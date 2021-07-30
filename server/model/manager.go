package model

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/blend/go-sdk/crypto"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/db/dbutil"
	"github.com/blend/go-sdk/ex"

	"github.com/blend/go-sdk/uuid"
)

// Manager is the common entrypoint for model functions.
type Manager struct {
	dbutil.BaseManager
}

// GetContentRatingByName gets a content rating by name.
func (m Manager) GetContentRatingByName(ctx context.Context, name string) (*ContentRating, error) {
	var rating ContentRating
	err := m.Invoke(ctx).Query(
		`SELECT * from content_rating where name = $1`, name,
	).Out(&rating)
	return &rating, err
}

// GetAllErrorsWithLimitAndOffset gets all the errors up to a limit.
func (m Manager) GetAllErrorsWithLimitAndOffset(ctx context.Context, limit, offset int) ([]Error, error) {
	var errors []Error
	err := m.Invoke(ctx).Query(`SELECT * FROM error ORDER BY created_utc desc LIMIT $1 OFFSET $2`, limit, offset).OutMany(&errors)
	return errors, err
}

// GetAllErrorsSince gets all the errors since a cutoff.
func (m Manager) GetAllErrorsSince(ctx context.Context, since time.Time) ([]Error, error) {
	var errors []Error
	err := m.Invoke(ctx).Query(`SELECT * FROM error WHERE created_utc > $1 ORDER BY created_utc desc`, since).OutMany(&errors)
	return errors, err
}

// GetAllImages returns all the images in the database.
func (m Manager) GetAllImages(ctx context.Context) ([]Image, error) {
	return m.GetImagesByID(ctx, nil)
}

// GetAllImagesWithContentRating gets all censored images
func (m Manager) GetAllImagesWithContentRating(ctx context.Context, contentRating int) ([]Image, error) {
	var imageIDs []imageSignature
	query := `select id from image where image.content_rating = $1`
	err := m.Invoke(ctx).Query(query, contentRating).OutMany(&imageIDs)

	if err != nil {
		return nil, err
	}
	images, err := m.GetImagesByID(ctx, imageSignatures(imageIDs).AsInt64s())
	if err != nil {
		return nil, err
	}
	return images, nil
}

// GetRandomImages returns an image by uuid.
func (m Manager) GetRandomImages(ctx context.Context, count int) ([]Image, error) {
	var imageIDs []imageSignature
	err := m.Invoke(ctx).Query(`select id from (select id, row_number() over (order by gen_random_uuid()) as rank from image where content_rating < 5) data where rank <= $1`, count).OutMany(&imageIDs)

	if err != nil {
		return nil, err
	}

	images, err := m.GetImagesByID(ctx, imageSignatures(imageIDs).AsInt64s())
	if err != nil {
		return nil, err
	}
	return images, err
}

// GetImageByID returns an image for an id.
func (m Manager) GetImageByID(ctx context.Context, id int64) (*Image, error) {
	images, err := m.GetImagesByID(ctx, []int64{id})
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return &Image{}, nil
	}
	return &images[0], err
}

// GetImageByUUID returns an image by uuid.
func (m Manager) GetImageByUUID(ctx context.Context, uuid string) (*Image, error) {
	var image imageSignature
	err := m.Invoke(ctx).Query(`select id from image where uuid = $1`, uuid).Out(&image)
	if err != nil {
		return nil, err
	}

	images, err := m.GetImagesByID(ctx, []int64{image.ID})
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return &Image{}, nil
	}

	return &images[0], err
}

// GetImageByMD5 returns an image by uuid.
func (m Manager) GetImageByMD5(ctx context.Context, md5sum []byte) (*Image, error) {
	image := Image{}
	imageColumns := db.Columns(Image{}).ColumnNames()
	err := m.Invoke(ctx).Query(fmt.Sprintf(`select %s from image where md5 = $1`, strings.Join(imageColumns, ",")), md5sum).Out(&image)
	return &image, err
}

// UpdateImageDisplayName sets just the display name for an image.
func (m Manager) UpdateImageDisplayName(ctx context.Context, imageID int64, displayName string) error {
	return m.Invoke(ctx).Exec("update image set display_name = $2 where id = $1", imageID, displayName)
}

// DeleteImageByID deletes an image fully.
func (m Manager) DeleteImageByID(ctx context.Context, imageID int64) error {
	err := m.Invoke(ctx).Exec(`delete from vote_summary where image_id = $1`, imageID)
	if err != nil {
		return err
	}
	err = m.Invoke(ctx).Exec(`delete from vote where image_id = $1`, imageID)
	if err != nil {
		return err
	}
	return m.Invoke(ctx).Exec(`delete from image where id = $1`, imageID)
}

func (m Manager) searchImagesInternal(ctx context.Context, query string, excludeUUIDs []string, contentRatingFilter int) ([]imageSignature, error) {
	var imageIDs []imageSignature

	args := []interface{}{
		query,
		contentRatingFilter,
	}

	var excludedClause string
	if len(excludeUUIDs) > 0 {
		excludedClause = fmt.Sprintf("and i.uuid not in (%s)", db.ParamTokens(len(args)+1, len(excludeUUIDs)))
		for _, excluded := range excludeUUIDs {
			args = append(args, excluded)
		}
	}

	searchImageQuery := fmt.Sprintf(`
	select
		vs.image_id as id
		, sum(ts.score * vs.votes_total) as score
	from
		(
			select
				t.id as tag_id
				, similarity(t.tag_value, $1) as score
			from
				tag t
			where
				similarity(t.tag_value, $1) > show_limit()
		) ts
		join vote_summary vs on vs.tag_id = ts.tag_id
		join image i on vs.image_id = i.id
	where
		vs.votes_total > 0
		and i.content_rating <= $2
		%s
	group by
		vs.image_id
	order by
		score desc;
	`, excludedClause)

	err := m.Invoke(ctx).Query(searchImageQuery, args...).OutMany(&imageIDs)
	return imageIDs, err
}

// SearchImages searches for an image.
func (m Manager) SearchImages(ctx context.Context, query string, contentRatingFilter int) ([]Image, error) {
	imageIDs, err := m.searchImagesInternal(ctx, query, nil, contentRatingFilter)
	if err != nil {
		return nil, err
	}

	if len(imageIDs) == 0 {
		return []Image{}, nil
	}

	ids := imageSignatures(imageIDs).AsInt64s()
	return m.GetImagesByID(ctx, ids)
}

// SearchImagesWeightedRandom pulls a random count of images based on a search query. The most common `count` is 1.
func (m Manager) SearchImagesWeightedRandom(ctx context.Context, query string, contentRatingFilter, count int) ([]Image, error) {
	imageIDs, err := m.searchImagesInternal(ctx, query, nil, contentRatingFilter)
	if err != nil {
		return nil, err
	}
	if len(imageIDs) == 0 {
		return []Image{}, nil
	}

	var finalImages []imageSignature

	if len(imageIDs) == 1 {
		finalImages = imageIDs
	} else {
		finalImages = imageSignatures(imageIDs).WeightedRandom(count)
	}

	return m.GetImagesByID(ctx, imageSignatures(finalImages).AsInt64s())
}

// SearchImagesBestResult pulls the best result for a query.
// This method is used for slack searches.
func (m Manager) SearchImagesBestResult(ctx context.Context, query string, excludeUUIDs []string, contentRating int) (*Image, error) {
	imageIDs, err := m.searchImagesInternal(ctx, query, excludeUUIDs, contentRating)
	if err != nil {
		return nil, err
	}
	if len(imageIDs) == 0 {
		return nil, nil
	}

	var imagesWithBestScore []imageSignature
	if len(imageIDs) == 1 {
		imagesWithBestScore = imageIDs
	} else {
		var bestScore float64
		for _, i := range imageIDs {
			if i.Score > bestScore {
				bestScore = i.Score
			}
		}

		for _, i := range imageIDs {
			if i.Score == bestScore {
				imagesWithBestScore = append(imagesWithBestScore, i)
			}
		}
	}

	bestImageID := imageSignatures(imagesWithBestScore).WeightedRandom(1).AsInt64s()

	images, err := m.GetImagesByID(ctx, bestImageID)

	if err != nil {
		return nil, err
	}

	if len(images) > 0 {
		return &images[0], nil
	}

	return nil, nil
}

// GetImagesForUserID returns images for a user.
func (m Manager) GetImagesForUserID(ctx context.Context, userID int64) ([]Image, error) {
	var imageIDs []imageSignature
	imageQuery := `select i.id from image i where created_by = $1`
	err := m.Invoke(ctx).Query(imageQuery, userID).OutMany(&imageIDs)
	if err != nil {
		return nil, err
	}

	if len(imageIDs) == 0 {
		return []Image{}, nil
	}

	ids := imageSignatures(imageIDs).AsInt64s()
	return m.GetImagesByID(ctx, ids)
}

// GetImagesByID returns images with tags for a list of ids.
func (m Manager) GetImagesByID(ctx context.Context, ids []int64) ([]Image, error) {
	var err error
	var populateErr error

	imageColumns := db.Columns(Image{}).ColumnNames()

	imageQueryAll := fmt.Sprintf(`select %s from image`, strings.Join(imageColumns, ","))
	imageQuerySingle := fmt.Sprintf(`%s where id = $1`, imageQueryAll)
	imageQueryMany := fmt.Sprintf(`%s where id = ANY($1::bigint[])`, imageQueryAll)

	tagQueryAll := `
	select
		t.*
		, u.uuid as created_by_uuid
		, vs.image_id
		, vs.votes_for
		, vs.votes_against
		, vs.votes_total
		, row_number() over (partition by image_id order by vs.votes_total desc) as vote_rank
	from
				tag t
		join 	vote_summary 	vs 	on vs.tag_id = t.id
		join 	users 			u 	on u.id = t.created_by
	`
	tagQuerySingle := fmt.Sprintf(`%s where vs.image_id = $1`, tagQueryAll)
	tagQueryMany := fmt.Sprintf(`%s where vs.image_id = ANY($1::bigint[])`, tagQueryAll)

	tagQueryOuter := `
		select * from (
		%s
		) as intermediate
		where vote_rank <= 5
	`

	userQueryAll := `select u.* from image i join users u on i.created_by = u.id`
	userQuerySingle := fmt.Sprintf(`%s where i.id = $1`, userQueryAll)
	userQueryMany := fmt.Sprintf(`%s where i.id = ANY($1::bigint[])`, userQueryAll)

	intermediateImages := []*Image{}
	imageLookup := map[int64]*Image{}
	userLookup := map[int64]*User{}

	imageConsumer := func(r db.Rows) error {
		i := &Image{}
		populateErr = i.Populate(r)
		if populateErr != nil {
			return populateErr
		}
		intermediateImages = append(intermediateImages, i)
		imageLookup[i.ID] = i
		return nil
	}

	tagConsumer := func(r db.Rows) error {
		t := &Tag{}
		populateErr = t.PopulateExtra(r)
		if populateErr != nil {
			return populateErr
		}

		i := imageLookup[t.ImageID]
		if i != nil {
			i.Tags = append(i.Tags, *t)
		}

		return nil
	}

	userConsumer := func(r db.Rows) error {
		u := &User{}
		populateErr = u.Populate(r)
		if populateErr != nil {
			return populateErr
		}
		userLookup[u.ID] = u
		return nil
	}

	if len(ids) > 1 {
		idsCSV := fmt.Sprintf("{%s}", csvOfInt(ids))
		err = m.Invoke(ctx).Query(imageQueryMany, idsCSV).Each(imageConsumer)
		if err != nil {
			return nil, err
		}
		err = m.Invoke(ctx).Query(fmt.Sprintf(tagQueryOuter, tagQueryMany), idsCSV).Each(tagConsumer)
		if err != nil {
			return nil, err
		}
		err = m.Invoke(ctx).Query(userQueryMany, idsCSV).Each(userConsumer)
		if err != nil {
			return nil, err
		}
	} else if len(ids) == 1 {
		err = m.Invoke(ctx).Query(imageQuerySingle, ids[0]).Each(imageConsumer)
		if err != nil {
			return nil, err
		}
		err = m.Invoke(ctx).Query(fmt.Sprintf(tagQueryOuter, tagQuerySingle), ids[0]).Each(tagConsumer)
		if err != nil {
			return nil, err
		}
		err = m.Invoke(ctx).Query(userQuerySingle, ids[0]).Each(userConsumer)
		if err != nil {
			return nil, err
		}
	} else {
		err = m.Invoke(ctx).Query(imageQueryAll).Each(imageConsumer)
		if err != nil {
			return nil, err
		}
		err = m.Invoke(ctx).Query(fmt.Sprintf(tagQueryOuter, tagQueryAll)).Each(tagConsumer)
		if err != nil {
			return nil, err
		}
		err = m.Invoke(ctx).Query(userQueryAll).Each(userConsumer)
		if err != nil {
			return nil, err
		}
	}

	finalImages := make([]Image, len(intermediateImages))
	for x := 0; x < len(intermediateImages); x++ {
		img := intermediateImages[x]
		if u, ok := userLookup[img.CreatedBy]; ok {
			img.CreatedByUser = u
		}

		finalImages[x] = *img
	}

	if len(ids) > 1 {
		sort.Sort(newImagesByIndex(&finalImages, ids))
	}

	return finalImages, nil
}

// SetVoteSummaryVoteCounts updates the votes for a tag to an image.
func (m Manager) SetVoteSummaryVoteCounts(ctx context.Context, imageID, tagID int64, votesFor, votesAgainst int) error {
	votesTotal := votesFor - votesAgainst
	return m.Invoke(ctx).Exec(`update vote_summary vs set votes_for = $1, votes_against = $2, votes_total = $3 where image_id = $4 and tag_id = $5`, votesFor, votesAgainst, votesTotal, imageID, tagID)
}

// SetVoteSummaryTagID re-assigns a vote summary's tag.
func (m Manager) SetVoteSummaryTagID(ctx context.Context, imageID, oldTagID, newTagID int64) error {
	return m.Invoke(ctx).Exec(`update vote_summary set tag_id = $1 where image_id = $2 and tag_id = $3`, newTagID, imageID, oldTagID)
}

// CreateOrUpdateVote votes for a tag for an image in the db.
func (m Manager) CreateOrUpdateVote(ctx context.Context, userID, imageID, tagID int64, isUpvote bool) (bool, error) {
	existing, existingErr := m.GetVoteSummary(ctx, imageID, tagID)
	if existingErr != nil {
		return false, existingErr
	}

	if existing.IsZero() {
		itv := NewVoteSummary(imageID, tagID, userID, time.Now().UTC())
		if isUpvote {
			itv.VotesFor = 1
			itv.VotesAgainst = 0
			itv.VotesTotal = 1
		} else {
			itv.VotesFor = 0
			itv.VotesAgainst = 1
			itv.VotesTotal = -1
		}
		err := m.Invoke(ctx).Create(itv)
		if err != nil {
			return true, err
		}
	} else {
		//check if user has already voted for this image ...
		if isUpvote {
			existing.VotesFor = existing.VotesFor + 1
		} else {
			existing.VotesAgainst = existing.VotesAgainst + 1
		}
		existing.LastVoteBy = userID
		existing.LastVoteUTC = time.Now().UTC()
		existing.VotesTotal = existing.VotesFor - existing.VotesAgainst

		updateErr := m.Invoke(ctx).Update(existing)
		if updateErr != nil {
			return false, updateErr
		}
	}

	err := m.DeleteVote(ctx, userID, imageID, tagID)
	if err != nil {
		return false, err
	}

	logEntry := NewVote(userID, imageID, tagID, isUpvote)
	return false, m.Invoke(ctx).Create(logEntry)
}

func (m Manager) getVoteSummaryQuery(whereClause string) string {
	return fmt.Sprintf(`
select
	vs.*
	, i.uuid as image_uuid
	, t.uuid as tag_uuid
	, u.uuid as last_vote_by_uuid
from
	vote_summary vs
	join image i on i.id = vs.image_id
	join tag t on t.id = vs.tag_id
	join users u on u.id = vs.last_vote_by
%s
`, whereClause)
}

// GetVoteSummariesForImage grabs all vote counts for an image (i.e. for all the tags).
func (m Manager) GetVoteSummariesForImage(ctx context.Context, imageID int64) ([]VoteSummary, error) {
	summaries := []VoteSummary{}
	err := m.Invoke(ctx).Query(m.getVoteSummaryQuery("where image_id = $1"), imageID).OutMany(&summaries)
	return summaries, err
}

// GetVoteSummariesForTag grabs all vote counts for an image (i.e. for all the tags).
func (m Manager) GetVoteSummariesForTag(ctx context.Context, tagID int64) ([]VoteSummary, error) {
	summaries := []VoteSummary{}
	err := m.Invoke(ctx).Query(m.getVoteSummaryQuery("where tag_id = $1"), tagID).OutMany(&summaries)
	return summaries, err
}

// GetVoteSummary fetches an VoteSummary by constituent pks.
func (m Manager) GetVoteSummary(ctx context.Context, imageID, tagID int64) (*VoteSummary, error) {
	var imv VoteSummary
	query := `select * from vote_summary where image_id = $1 and tag_id = $2`
	err := m.Invoke(ctx).Query(query, imageID, tagID).Out(&imv)
	return &imv, err
}

// GetImagesForTagID gets all the images attached to a tag.
func (m Manager) GetImagesForTagID(ctx context.Context, tagID int64) ([]Image, error) {
	var imageIDs []imageSignature
	err := m.Invoke(ctx).Query(`select image_id as id from vote_summary vs where tag_id = $1 order by vs.votes_total desc;`, tagID).OutMany(&imageIDs)
	if err != nil {
		return nil, err
	}

	return m.GetImagesByID(ctx, imageSignatures(imageIDs).AsInt64s())
}

// GetTagsForImageID returns all the tags for an image.
func (m Manager) GetTagsForImageID(ctx context.Context, imageID int64) ([]Tag, error) {
	var tags []Tag
	query := `
select
	t.*
    , u.uuid as created_by_uuid
	, vs.image_id
	, vs.votes_for
	, vs.votes_against
	, vs.votes_total
    , row_number() over (partition by vs.image_id order by vs.votes_total desc) as vote_rank
from
	tag t
	join vote_summary vs on t.id = vs.tag_id
    join users u on u.id = t.created_by
where
	vs.image_id = $1
order by
	vs.votes_total desc;
`
	err := m.Invoke(ctx).Query(query, imageID).Each(func(r db.Rows) error {
		t := &Tag{}
		popErr := t.PopulateExtra(r)
		if popErr != nil {
			return popErr
		}
		tags = append(tags, *t)
		return nil
	})
	return tags, err
}

// ReconcileVoteSummaryTotals queries the `vote` table to fill the vote count aggregate columns on `vote_summary`.
func (m Manager) ReconcileVoteSummaryTotals(ctx context.Context, imageID, tagID int64) error {
	query := `
update
	vote_summary
set
	votes_for = data.votes_for
	, votes_against = data.votes_against
	, votes_total = data.votes_total
from
(
	select
		coalesce(sum(votes_for), 0) as votes_for
		, coalesce(sum(votes_against), 0) as votes_against
		, coalesce(sum(votes_for), 0) - coalesce(sum(votes_against),0) as votes_total
	from (
		select
			case when is_upvote = true then 1 else 0 end as votes_for
			, case when is_upvote = true then 0 else 1 end as votes_against
		from
			vote v
		where
			v.image_id = $1
			and v.tag_id = $2
	) as data_inner
) data
where
	vote_summary.image_id = $1
	and vote_summary.tag_id = $2
`
	return m.Invoke(ctx).Exec(query, imageID, tagID)
}

// DeleteVoteSummary deletes an association between an image and a tag.
func (m Manager) DeleteVoteSummary(ctx context.Context, imageID, tagID int64) error {
	return m.Invoke(ctx).Exec(`delete from vote_summary where image_id = $1 and tag_id = $2`, imageID, tagID)
}

// GetAllTags returns all the tags in the db.
func (m Manager) GetAllTags(ctx context.Context) ([]Tag, error) {
	all := []Tag{}
	err := m.Invoke(ctx).GetAll(&all)
	return all, err
}

// GetRandomTags gets a random selection of tags.
func (m Manager) GetRandomTags(ctx context.Context, count int) ([]Tag, error) {
	tags := []Tag{}
	err := m.Invoke(ctx).Query(`select * from tag order by gen_random_uuid() limit $1;`, count).OutMany(&tags)
	return tags, err
}

// GetTagByID returns a tag for a id.
func (m Manager) GetTagByID(ctx context.Context, id int64) (*Tag, error) {
	tag := Tag{}
	err := m.Invoke(ctx).Query(`select * from tag where id = $1`, id).Out(&tag)
	return &tag, err
}

// GetTagByUUID returns a tag for a uuid.
func (m Manager) GetTagByUUID(ctx context.Context, uuid string) (*Tag, error) {
	tag := Tag{}
	err := m.Invoke(ctx).Query(`select * from tag where uuid = $1`, uuid).Out(&tag)
	return &tag, err
}

// GetTagByValue returns a tag for a uuid.
func (m Manager) GetTagByValue(ctx context.Context, tagValue string) (*Tag, error) {
	tag := Tag{}
	err := m.Invoke(ctx).Query(`select * from tag where tag_value ilike $1`, tagValue).Out(&tag)
	return &tag, err
}

// SearchTags searches tags.
func (m Manager) SearchTags(ctx context.Context, query string) ([]Tag, error) {
	tags := []Tag{}
	err := m.Invoke(ctx).Query(`select * from tag where tag_value % $1 order by similarity(tag_value, $1) desc;`, query).OutMany(&tags)
	return tags, err
}

// SearchTagsRandom searches tags taking a randomly selected count.
func (m Manager) SearchTagsRandom(ctx context.Context, query string, count int) ([]Tag, error) {
	tags := []Tag{}
	err := m.Invoke(ctx).Query(`select * from tag where tag_value % $1 order by gen_random_uuid() limit $2;`, query, count).OutMany(&tags)
	return tags, err
}

// SetTagValue sets a tag value
func (m Manager) SetTagValue(ctx context.Context, tagID int64, tagValue string) error {
	return m.Invoke(ctx).Exec(`update tag set tag_value = $1 where id = $2`, tagValue, tagID)
}

// DeleteTagByID deletes a tag.
func (m Manager) DeleteTagByID(ctx context.Context, tagID int64) error {
	return m.Invoke(ctx).Exec(`delete from tag where id = $1`, tagID)
}

// DeleteTagAndVotesByID deletes an tag fully.
func (m Manager) DeleteTagAndVotesByID(ctx context.Context, tagID int64) error {
	err := m.Invoke(ctx).Exec(`delete from vote_summary where tag_id = $1`, tagID)
	if err != nil {
		return err
	}
	err = m.Invoke(ctx).Exec(`delete from vote where tag_id = $1`, tagID)
	if err != nil {
		return err
	}
	return m.DeleteTagByID(ctx, tagID)
}

// MergeTags merges the fromTagID into the toTagID, deleting the fromTagID.
func (m Manager) MergeTags(ctx context.Context, fromTagID, toTagID int64) error {
	votes, err := m.GetVotesForTag(ctx, fromTagID)
	if err != nil {
		return err
	}

	for _, vote := range votes {
		existingVote, err := m.GetVote(ctx, vote.UserID, vote.ImageID, toTagID)
		if err != nil {
			return err
		}

		if existingVote.IsZero() {
			err = m.SetVoteTagID(ctx, vote.UserID, vote.ImageID, fromTagID, toTagID)
			if err != nil {
				return err
			}
		} else {
			err = m.DeleteVote(ctx, vote.UserID, vote.ImageID, fromTagID)
			if err != nil {
				return err
			}
		}
	}

	links, err := m.GetVoteSummariesForTag(ctx, fromTagID)
	if err != nil {
		return err
	}

	for _, link := range links {
		existingLink, err := m.GetVoteSummary(ctx, link.ImageID, toTagID)
		if err != nil {
			return err
		}

		if existingLink.IsZero() {
			err = m.SetVoteSummaryTagID(ctx, link.ImageID, fromTagID, toTagID)
			if err != nil {
				return err
			}
		} else {
			err = m.ReconcileVoteSummaryTotals(ctx, link.ImageID, toTagID)
			if err != nil {
				return err
			}

			err = m.DeleteVoteSummary(ctx, link.ImageID, fromTagID)
			if err != nil {
				return err
			}
		}
	}

	return m.DeleteTagByID(ctx, fromTagID)
}

// DeleteOrphanedTags deletes tags that have no vote_summary link to an image.
func (m Manager) DeleteOrphanedTags(ctx context.Context) error {
	err := m.Invoke(ctx).Exec(`delete from vote where not exists (select 1 from vote_summary vs where vs.tag_id = vote.tag_id);`)
	if err != nil {
		return err
	}
	return m.Invoke(ctx).Exec(`delete from tag where not exists (select 1 from vote_summary vs where vs.tag_id = tag.id);`)
}

// GetModerationForUserID gets all the moderation entries for a user.
func (m Manager) GetModerationForUserID(ctx context.Context, userID int64) ([]Moderation, error) {
	var moderationLog []Moderation
	whereClause := `where user_id = $1`
	err := m.Invoke(ctx).Query(m.getModerationQuery(whereClause), userID).Each(m.moderationConsumer(&moderationLog))
	return moderationLog, err
}

// GetModerationsByTime returns all moderation entries after a specific time.
func (m Manager) GetModerationsByTime(ctx context.Context, after time.Time) ([]Moderation, error) {
	var moderationLog []Moderation
	whereClause := `where timestamp_utc > $1`
	err := m.Invoke(ctx).Query(m.getModerationQuery(whereClause), after).Each(m.moderationConsumer(&moderationLog))
	return moderationLog, err
}

// GetModerationLogByCountAndOffset returns all moderation entries after a specific time.
func (m Manager) GetModerationLogByCountAndOffset(ctx context.Context, count, offset int) ([]Moderation, error) {
	var moderationLog []Moderation
	query := m.getModerationQuery("")
	query = query + `limit $1 offset $2`
	err := m.Invoke(ctx).Query(query, count, offset).Each(m.moderationConsumer(&moderationLog))
	return moderationLog, err
}

func (m Manager) getModerationQuery(whereClause string) string {
	moderatorColumns := db.Columns(User{}).NotReadOnly().CopyWithColumnPrefix("moderator_").ColumnNamesCSVFromAlias("mu")
	userColumns := db.Columns(User{}).NotReadOnly().CopyWithColumnPrefix("target_user_").ColumnNamesCSVFromAlias("u")
	imageColumns := db.Columns(Image{}).NotReadOnly().CopyWithColumnPrefix("image_").ColumnNamesCSVFromAlias("i")
	tagColumns := db.Columns(Tag{}).NotReadOnly().CopyWithColumnPrefix("tag_").ColumnNamesCSVFromAlias("t")

	return fmt.Sprintf(`
	select
		m.*,
		%s,
		%s,
		%s,
		%s
	from
		moderation m
		join users mu on m.user_id = mu.id
		left join users u on m.noun = u.uuid or m.secondary_noun = u.uuid
		left join image i on m.noun = i.uuid or m.secondary_noun = i.uuid
		left join tag t on m.noun = t.uuid or m.secondary_noun = t.uuid
	%s
	order by timestamp_utc desc
	`,
		moderatorColumns,
		userColumns,
		imageColumns,
		tagColumns,
		whereClause)
}

func (m Manager) moderationConsumer(moderationLog *[]Moderation) db.RowsConsumer {
	moderationColumns := db.Columns(Moderation{})
	moderatorColumns := db.Columns(User{}).NotReadOnly().CopyWithColumnPrefix("moderator_")
	userColumns := db.Columns(User{}).NotReadOnly().CopyWithColumnPrefix("target_user_")
	imageColumns := db.Columns(Image{}).NotReadOnly().CopyWithColumnPrefix("image_")
	tagColumns := db.Columns(Tag{}).NotReadOnly().CopyWithColumnPrefix("tag_")

	return func(r db.Rows) error {
		var m Moderation
		var mu User

		var u User
		var i Image
		var t Tag

		err := db.PopulateByName(&m, r, moderationColumns)
		if err != nil {
			return err
		}

		err = db.PopulateByName(&mu, r, moderatorColumns)
		if err != nil {
			return err
		}
		m.Moderator = &mu

		err = db.PopulateByName(&u, r, userColumns)
		if err == nil && !u.IsZero() {
			m.User = &u
		}

		err = db.PopulateByName(&i, r, imageColumns)
		if err == nil && !i.IsZero() {
			m.Image = &i
		}

		err = db.PopulateByName(&t, r, tagColumns)
		if err == nil && !t.IsZero() {
			m.Tag = &t
		}

		*moderationLog = append(*moderationLog, m)
		return nil
	}
}

// GetAllUsers returns all the users.
func (m Manager) GetAllUsers(ctx context.Context) ([]User, error) {
	var all []User
	err := m.Invoke(ctx).GetAll(&all)
	return all, err
}

// GetUsersByCountAndOffset returns users by count and offset.
func (m Manager) GetUsersByCountAndOffset(ctx context.Context, count, offset int) ([]User, error) {
	var all []User
	err := m.Invoke(ctx).Query(`select * from users order by created_utc desc limit $1 offset $2`, count, offset).OutMany(&all)
	return all, err
}

// GetUserByID returns a user by id.
func (m Manager) GetUserByID(ctx context.Context, id int64) (*User, error) {
	var user User
	err := m.Invoke(ctx).Get(&user, id)
	return &user, err
}

// GetUserByUUID returns a user for a uuid.
func (m Manager) GetUserByUUID(ctx context.Context, uuid string) (*User, error) {
	var user User
	err := m.Invoke(ctx).Query(`select * from users where uuid = $1`, uuid).Out(&user)
	return &user, err
}

// SearchUsers searches users by searchString.
func (m Manager) SearchUsers(ctx context.Context, query string) ([]User, error) {
	var users []User
	queryFormat := fmt.Sprintf("%%%s%%", query)
	sqlQuery := `select * from users where username ilike $1 or first_name ilike $1 or last_name ilike $1 or email_address ilike $1`
	err := m.Invoke(ctx).Query(sqlQuery, queryFormat).OutMany(&users)
	return users, err
}

// GetUserByUsername returns a user for a uuid.
func (m Manager) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := m.Invoke(ctx).Query(`select * from users where username = $1`, username).Out(&user)
	return &user, err
}

func (m Manager) createSearchHistoryQuery(whereClause ...string) string {
	searchColumns := db.Columns(SearchHistory{}).NotReadOnly().ColumnNamesCSVFromAlias("sh")
	imageColumns := db.Columns(Image{}).NotReadOnly().CopyWithColumnPrefix("image_").ColumnNamesCSVFromAlias("i")
	tagColumns := db.Columns(Tag{}).NotReadOnly().CopyWithColumnPrefix("tag_").ColumnNamesCSVFromAlias("t")

	query := `
	select
		%s,
		%s,
		%s
	from
		search_history sh
		left join image i on i.id = sh.image_id
		left join tag t on t.id = sh.tag_id
	%s
	order by timestamp_utc desc
	`
	if len(whereClause) > 0 {
		return fmt.Sprintf(query, searchColumns, imageColumns, tagColumns, whereClause[0])
	}
	return fmt.Sprintf(query, searchColumns, imageColumns, tagColumns, "")
}

func (m Manager) searchHistoryConsumer(searchHistory *[]SearchHistory) db.RowsConsumer {
	searchColumns := db.Columns(SearchHistory{}).NotReadOnly()
	imageColumns := db.Columns(Image{}).NotReadOnly().CopyWithColumnPrefix("image_")
	tagColumns := db.Columns(Tag{}).NotReadOnly().CopyWithColumnPrefix("tag_")

	return func(r db.Rows) error {
		var sh SearchHistory
		var i Image
		var t Tag

		err := db.PopulateByName(&sh, r, searchColumns)
		if err != nil {
			return err
		}

		err = db.PopulateByName(&i, r, imageColumns)
		if err == nil && !i.IsZero() {
			sh.Image = &i
		}

		err = db.PopulateByName(&t, r, tagColumns)
		if err == nil && !t.IsZero() {
			sh.Tag = &t
		}

		*searchHistory = append(*searchHistory, sh)
		return nil
	}
}

// GetSearchHistory returns the entire search history in chrono order.
func (m Manager) GetSearchHistory(ctx context.Context) ([]SearchHistory, error) {
	var searchHistory []SearchHistory
	query := m.createSearchHistoryQuery()
	err := m.Invoke(ctx).Query(query).Each(m.searchHistoryConsumer(&searchHistory))
	return searchHistory, err
}

// GetSearchHistoryByCountAndOffset returns the search history in chrono order by count and offset.
func (m Manager) GetSearchHistoryByCountAndOffset(ctx context.Context, count, offset int) ([]SearchHistory, error) {
	var searchHistory []SearchHistory
	query := m.createSearchHistoryQuery()
	query = query + `limit $1 offset $2`
	err := m.Invoke(ctx).Query(query, count, offset).Each(m.searchHistoryConsumer(&searchHistory))
	return searchHistory, err
}

// GetAllSlackTeams gets all slack teams.
func (m Manager) GetAllSlackTeams(ctx context.Context) ([]SlackTeam, error) {
	var teams []SlackTeam
	err := m.Invoke(ctx).Query(`select * from slack_team order by team_name asc`).OutMany(&teams)
	return teams, err
}

// GetSlackTeamByTeamID gets a slack team by the team id.
func (m Manager) GetSlackTeamByTeamID(ctx context.Context, teamID string) (*SlackTeam, error) {
	var team SlackTeam
	err := m.Invoke(ctx).Get(&team, teamID)
	return &team, err
}

// GetUserAuthByToken returns an auth entry for the given auth token.
func (m Manager) GetUserAuthByToken(ctx context.Context, token string, key []byte) (*UserAuth, error) {
	if len(key) == 0 {
		return nil, ex.New("`ENCRYPTION_KEY` is not set, cannot continue.")
	}

	authTokenHash := crypto.HMAC512(key, []byte(token))

	var auth UserAuth
	err := m.Invoke(ctx).Query("SELECT * FROM user_auth where auth_token_hash = $1", authTokenHash).Out(&auth)
	return &auth, err
}

// GetUserAuthByProvider returns an auth entry for the given auth token.
func (m Manager) GetUserAuthByProvider(ctx context.Context, userID int64, provider string) (*UserAuth, error) {
	var auth UserAuth
	err := m.Invoke(ctx).Query("SELECT * FROM user_auth where user_id = $1 and provider = $2", userID, provider).Out(&auth)
	return &auth, err
}

// DeleteUserAuthForProvider deletes auth entries for a provider for a user.
func (m Manager) DeleteUserAuthForProvider(ctx context.Context, userID int64, provider string) error {
	return m.Invoke(ctx).Exec("DELETE FROM user_auth where user_id = $1 and provider = $2", userID, provider)
}

// DeleteUserSession removes a session from the db.
func (m Manager) DeleteUserSession(ctx context.Context, userID int64, sessionID string) error {
	return m.Invoke(ctx).Exec("DELETE FROM user_session where user_id = $1 and session_id = $2", userID, sessionID)
}

// GetVotesForUser gets all the vote log entries for a user.
func (m Manager) GetVotesForUser(ctx context.Context, userID int64) ([]Vote, error) {
	votes := []Vote{}
	err := m.Invoke(ctx).Query(m.getVotesQuery("where v.user_id = $1"), userID).OutMany(&votes)
	return votes, err
}

// GetVotesForImage gets all the votes log entries for an image.
func (m Manager) GetVotesForImage(ctx context.Context, imageID int64) ([]Vote, error) {
	votes := []Vote{}
	err := m.Invoke(ctx).Query(m.getVotesQuery("where v.image_id = $1"), imageID).OutMany(&votes)
	return votes, err
}

// GetVotesForTag gets all the votes log entries for an image.
func (m Manager) GetVotesForTag(ctx context.Context, tagID int64) ([]Vote, error) {
	votes := []Vote{}
	err := m.Invoke(ctx).Query(m.getVotesQuery("where v.tag_id = $1"), tagID).OutMany(&votes)
	return votes, err
}

// GetVotesForUserForImage gets the votes for an image by a user.
func (m Manager) GetVotesForUserForImage(ctx context.Context, userID, imageID int64) ([]Vote, error) {
	votes := []Vote{}
	err := m.Invoke(ctx).Query(m.getVotesQuery("where v.user_id = $1 and v.image_id = $2"), userID, imageID).OutMany(&votes)
	return votes, err
}

// GetVotesForUserForTag gets the votes for an image by a user.
func (m Manager) GetVotesForUserForTag(ctx context.Context, userID, tagID int64) ([]Vote, error) {
	votes := []Vote{}
	err := m.Invoke(ctx).Query(m.getVotesQuery("where v.user_id = $1 and v.tag_id = $2"), userID, tagID).OutMany(&votes)
	return votes, err
}

// GetVote gets a user's vote for an image and a tag.
func (m Manager) GetVote(ctx context.Context, userID, imageID, tagID int64) (*Vote, error) {
	voteLog := Vote{}
	err := m.Invoke(ctx).Query(m.getVotesQuery("where v.user_id = $1 and v.image_id = $2 and v.tag_id = $3"), userID, imageID, tagID).Out(&voteLog)
	return &voteLog, err
}

// SetVoteTagID sets the tag_id for a vote object.
func (m Manager) SetVoteTagID(ctx context.Context, userID, imageID, oldTagID, newTagID int64) error {
	return m.Invoke(ctx).Exec(`update vote set tag_id = $1 where user_id = $2 and image_id = $3 and tag_id = $4`, newTagID, userID, imageID, oldTagID)
}

// DeleteVote deletes a vote.
func (m Manager) DeleteVote(ctx context.Context, userID, imageID, tagID int64) error {
	return m.Invoke(ctx).Exec(`DELETE from vote where user_id = $1 and image_id = $2 and tag_id = $3`, userID, imageID, tagID)
}

func (m Manager) getVotesQuery(whereClause string) string {
	return fmt.Sprintf(`
select
	v.*
	, u.uuid as user_uuid
	, i.uuid as image_uuid
	, t.uuid as tag_uuid
from
	vote v
	join users u on v.user_id = u.id
	join image i on v.image_id = i.id
	join tag t on v.tag_id = t.id
%s
order by
	v.created_utc desc;
`, whereClause)
}

// GetSearchesPerDay retrieves the number of searches per day.
func (m Manager) GetSearchesPerDay(ctx context.Context, since time.Time) ([]DayCount, error) {
	data := []DayCount{}
	err := m.Invoke(ctx).Query(`
	select
		date_part('year', timestamp_utc) as year
		, date_part('month', timestamp_utc) as month
		, date_part('day', timestamp_utc) as day
		, count(*) as count
		from search_history
	where
		timestamp_utc > $1
	group by
		date_part('year', timestamp_utc)
		, date_part('month', timestamp_utc)
		, date_part('day', timestamp_utc)
	order by
		date_part('year', timestamp_utc) asc
		, date_part('month', timestamp_utc) asc
		, date_part('day', timestamp_utc) asc
	`, since).OutMany(&data)

	return data, err
}

// GetTopSearchedImages gets all the stats for images.
func (m Manager) GetTopSearchedImages(ctx context.Context, limit int) ([]ImageStats, error) {
	var results []ImageStats
	return results, nil
}

// GetImageStats gets image stats.
func (m Manager) GetImageStats(ctx context.Context, imageID int64) (*ImageStats, error) {
	var results ImageStats
	query := `
	select
		i.id as image_id
		, (select sum(votes_total) from vote_summary where image_id = $1) as votes_total
		, (select count(image_id) from search_history where image_id = $1) as searches
	from
		image i
	where
		i.id = $1
	`

	err := m.Invoke(ctx).Query(query, imageID).Out(&results)
	if err != nil {
		return nil, err
	}

	return &results, nil
}

// GetSiteStats returns the stats for the site.
func (m Manager) GetSiteStats(ctx context.Context) (*SiteStats, error) {
	imageCountQuery := `select coalesce(count(*), 0) as value from image;`
	tagCountQuery := `select coalesce(count(*), 0) as value from tag;`
	userCountQuery := `select coalesce(count(*), 0) as value from users;`
	karmaTotalQuery := `select coalesce(sum(votes_total), 0) as value from vote_summary;`
	orphanedTagCountQuery := `select coalesce(count(*), 0) from tag t where not exists (select 1 from vote_summary vs where vs.tag_id = t.id);`

	var userCount int
	var imageCount int
	var tagCount int
	var karmaTotal int
	var orphanedTagCount int

	err := m.Invoke(ctx).Query(userCountQuery).Scan(&userCount)
	if err != nil {
		return nil, err
	}
	err = m.Invoke(ctx).Query(imageCountQuery).Scan(&imageCount)
	if err != nil {
		return nil, err
	}
	err = m.Invoke(ctx).Query(tagCountQuery).Scan(&tagCount)
	if err != nil {
		return nil, err
	}
	err = m.Invoke(ctx).Query(karmaTotalQuery).Scan(&karmaTotal)
	if err != nil {
		return nil, err
	}
	err = m.Invoke(ctx).Query(orphanedTagCountQuery).Scan(&orphanedTagCount)
	if err != nil {
		return nil, err
	}

	return &SiteStats{
		UserCount:        userCount,
		ImageCount:       imageCount,
		TagCount:         tagCount,
		KarmaTotal:       karmaTotal,
		OrphanedTagCount: orphanedTagCount,
	}, nil
}

//
// testing stuff
//

// CreateObject creates an object (for use with the work queue)
func (m Manager) CreateObject(ctx context.Context, state ...interface{}) error {
	for _, obj := range state {
		if err := m.Invoke(ctx).Create(obj); err != nil {
			return err
		}
	}
	return nil
}

// CreateTestUser creates a test user.
func (m Manager) CreateTestUser(ctx context.Context) (*User, error) {
	u := NewUser(fmt.Sprintf("__test_user_%s__", uuid.V4().String()))
	u.FirstName = "Test"
	u.LastName = "User"
	err := m.Invoke(ctx).Create(u)
	return u, err
}

// CreateTestTag creates a test tag.
func (m Manager) CreateTestTag(ctx context.Context, userID int64, tagValue string) (*Tag, error) {
	tag := NewTag(userID, tagValue)
	err := m.Invoke(ctx).Create(tag)
	return tag, err
}

// CreateTestTagForImageWithVote creates a test tag for an image and seeds the vote summary for it.
func (m Manager) CreateTestTagForImageWithVote(ctx context.Context, userID, imageID int64, tagValue string) (*Tag, error) {
	tag, err := m.CreateTestTag(ctx, userID, tagValue)
	if err != nil {
		return nil, err
	}

	existing, err := m.GetVoteSummary(ctx, imageID, tag.ID)
	if err != nil {
		return nil, err
	}

	if existing.IsZero() {
		v := NewVoteSummary(imageID, tag.ID, userID, time.Now().UTC())
		v.VotesFor = 1
		v.VotesAgainst = 0
		v.VotesTotal = 1
		err = m.Invoke(ctx).Create(v)
		if err != nil {
			return nil, err
		}
	}

	err = m.DeleteVote(ctx, userID, imageID, tag.ID)
	if err != nil {
		return nil, err
	}

	v := NewVote(userID, imageID, tag.ID, true)
	err = m.Invoke(ctx).Create(v)
	if err != nil {
		return nil, err
	}
	return tag, err
}

// CreateTestImage creates a test image.
func (m Manager) CreateTestImage(ctx context.Context, userID int64) (*Image, error) {
	i := NewImage()
	i.CreatedBy = userID
	i.Extension = "gif"
	i.Width = 720
	i.Height = 480
	i.S3Bucket = uuid.V4().String()
	i.S3Key = uuid.V4().String()
	i.MD5 = uuid.V4()
	i.DisplayName = "Test Image"
	err := m.Invoke(ctx).Create(i)
	return i, err
}

// CreatTestVoteSummaryWithVote creates a test vote summary with an accopanying vote.
func (m Manager) CreatTestVoteSummaryWithVote(ctx context.Context, imageID, tagID, userID int64, votesFor, votesAgainst int) (*VoteSummary, error) {
	newLink := NewVoteSummary(imageID, tagID, userID, time.Now().UTC())
	newLink.VotesFor = votesFor
	newLink.VotesTotal = votesAgainst
	newLink.VotesTotal = votesFor - votesAgainst
	err := m.Invoke(ctx).Create(newLink)

	if err != nil {
		return nil, err
	}

	err = m.DeleteVote(ctx, userID, imageID, tagID)
	if err != nil {
		return nil, err
	}

	v := NewVote(userID, imageID, tagID, true)
	err = m.Invoke(ctx).Create(v)
	if err != nil {
		return nil, err
	}

	return newLink, nil
}

// CreateTestUserAuth creates a test user auth.
func (m Manager) CreateTestUserAuth(ctx context.Context, userID int64, token, secret string, key []byte) (*UserAuth, error) {
	ua, err := NewUserAuth(userID, token, secret, key)
	if err != nil {
		return ua, err
	}
	ua.Provider = "test"
	err = m.Invoke(ctx).Create(ua)
	return ua, err
}

// CreateTestUserSession creates a test user session.
func (m Manager) CreateTestUserSession(ctx context.Context, userID int64) (*UserSession, error) {
	us := NewUserSession(userID)
	err := m.Invoke(ctx).Create(us)
	return us, err
}

// CreateTestSearchHistory creates a test search history entry.
func (m Manager) CreateTestSearchHistory(ctx context.Context, source, searchQuery string, imageID, tagID *int64) (*SearchHistory, error) {
	sh := NewSearchHistory(source, searchQuery, true, imageID, tagID)
	err := m.Invoke(ctx).Create(sh)
	return sh, err
}
