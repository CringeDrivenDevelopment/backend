import {Hono} from "hono";
import type {JwtVariables} from "hono/jwt";
import {jwt} from "hono/jwt";
import {getMetadata, search} from "./utils/api.ts";
import {dl} from "./utils/dl.ts";

const secret = Bun.env.JWT_SECRET;
if (secret === undefined || secret === '') {
    console.error('Missing JWT_SECRET');
    process.exit(1);
}

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
        })
    }
});

// request download
app.post('/api/dl', async (c) => {
    const id = c.req.query('id');
    if (id === '' || id === undefined) {
        return c.json({});
    }

    console.log(`Received dl request for id: ${id}`);

    const metadata = await getMetadata(id);

    await dl(id, metadata);

    return c.json({})
});

// handle files
app.get('/api/dl/:id/:file', async (c) => {
    const id = c.req.param('id');
    if (id === '' || id === undefined) {
        return c.json({});
    }

    const file = c.req.param('file');
    if (file === '' || file === undefined) {
        return c.json({});
    }

    // path traversal block
    file.replaceAll('..', '');

    const blob = Bun.file(`./dl/${id}/${file}`);
    // check if exists
    if (!await blob.exists()) {
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
})

export default app