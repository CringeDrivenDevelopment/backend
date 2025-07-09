import {Hono} from "hono";
import type {JwtVariables} from "hono/jwt";
import {jwt} from "hono/jwt";
import {getMetadata, search} from "./utils/api.ts";
import {archive, dl} from "./utils/dl.ts";

const secret = Bun.env.JWT_SECRET;
if (secret === undefined || secret === '') {
    console.error('Missing JWT_SECRET');
    process.exit(1);
}
const cobaltUrl = Bun.env.COBALT_URL ?? 'http://localhost:9000';

type Variables = JwtVariables;
const app = new Hono<{ Variables: Variables }>();

// allow only authenticated
app.use(
    '/api/*',
    jwt({
        secret: secret,
    })
);


// search yt music
app.get('/api/search', async (c) => {
    const query = c.req.query('query');
    if (query === '' || query === undefined) {
        return c.json([]);
    }

    console.log(`Received search request for query: ${decodeURI(query)}`);

    try {
        const result = await search(query);

        return c.json(result);
    } catch (error) {
        console.error(error);
        c.status(500);
        return c.json({
            error: error,
        });
    }
});

// create archive
app.post('/api/archive', async (c) => {
    const body = await c.req.json<{
        songs: string[];
    }>();

    if (!body.songs || !Array.isArray(body.songs) || body.songs.length === 0) {
        return c.json({ error: 'Invalid request format' }, 400);
    }

    const hash = Bun.hash(body.songs.join(), Number.MAX_SAFE_INTEGER).toString();
    const archiveName = `${hash}.zip`

    if (!await Bun.file(`./dl/${archiveName}`).exists()) {
        archive(body.songs, archiveName).catch(err => {
            console.error(err);
        });
    }

    return c.json({
        filename: archiveName,
    });
});

// request download
app.post('/api/dl', async (c) => {
    const id = c.req.query('id');
    if (id === '' || id === undefined) {
        return c.json({});
    }

    console.log(`Received dl request for id: ${id}`);

    const blob = Bun.file(`./dl/${id}/hls.m3u8`);
    // check if exists
    if (await blob.exists()) {
        return c.json({});
    }

    const metadata = await getMetadata(id);

    if (metadata.length > 1200) {
        c.status(403);
        return c.json({});
    }

    dl(cobaltUrl, id, metadata).catch(err => {
        console.error(err);
    });

    return c.json({});
});

// handle files
app.get('/api/dl/:id?/:file?', async (c) => {
    let id = c.req.param('id') || '';
    let file = c.req.param('file') || '';

    if (!file && id) {
        file = id;
        id = '';
    }

    // path traversal block
    file = file.replaceAll('..', '');
    const path = id !== undefined ? `./dl/${id}/${file}` : `./dl/${file}`;
    const blob = Bun.file(path);
    console.log(`Received download request for path: ${path}`);
    // check if exists
    if (!await blob.exists()) {
        if (file == 'hls.m3u8' && id !== undefined) {
            console.log(`Received dl request for id: ${id}`);

            const metadata = await getMetadata(id);

            if (metadata.length > 1200) {
                c.status(403);
                return c.json({});
            }

            dl(cobaltUrl, id, metadata).catch(err => {
                console.error(err);
            });
        }
        c.status(404);
        return c.json({});
    }

    const arrbuf = await blob.arrayBuffer();
    const buffer = Buffer.from(arrbuf);

    return c.body(buffer, {
        headers: {
            'Content-Type': 'application/octet-stream',
        }
    });
});

export default app