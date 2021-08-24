export default undefined

/**
 * @typedef Post
 * @property {bigint} id 
 * @property {string} content
 * @property {boolean} NSFW      
 * @property {string=} spoiler_of
 * @property {number} likes_count
 * @property {number} comments_count
 * @property {string|Date} create_at
 * @property {User=} user
 * @property {boolean} mine
 * @property {boolean} liked
 * @property {boolean} subscribed
 */


/**
 * @typedef User
 * @property {bigint} id
 * @property {string} username
 * @property {string} avatar_url
 */

/**
 * @typedef timelineItem
 * @property {bigint} id
 * @property {Post=} post
 */

/**
 * @typedef CreatePostInput
 * @property {string} content
 * @property {boolean=} NSFW 
 * @property {string=} spoilerOf
 */