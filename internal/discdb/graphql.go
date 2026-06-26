package discdb

const getDiscByContentHash = `
	query GetDiscByContentHash($hash: String) {
	  mediaItems(
		where: {
		  releases: {some: {discs: {some: {disc: {contentHash: {eq: $hash}}}}}}
		}
	  ) {
		nodes {
		  title
		  year
		  slug
		  type
		  releases(
			where: {discs: {some: {disc: {contentHash: {eq: $hash}}}}}
		  ) {
			slug
			locale
			year
			title
			discs(
			  where: {disc: {contentHash: {eq: $hash}}}
			) {
			  contentHash
			  index
			  name
			  format
			  slug
			  titles {
				index
				duration
				displaySize
				sourceFile
				size
				segmentMap
				item {
				  title
				  season
				  episode
				  type
				}
			  }
			}
		  }
		}
	  }
	}`
