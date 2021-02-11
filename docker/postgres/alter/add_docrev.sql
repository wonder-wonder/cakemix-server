alter table documentrevision add column revision integer;

UPDATE
  documentrevision AS t1
SET
  revision = seqnum
FROM (
  SELECT
    uuid, updatedat,
    ROW_NUMBER() OVER (partition by "uuid" ORDER BY updatedat ASC) AS seqnum
  FROM
    documentrevision
) AS t2
WHERE 
  t1.uuid = t2.uuid AND t1.updatedat=t2.updatedat
;

ALTER TABLE documentrevision ALTER COLUMN revision SET NOT NULL;

ALTER TABLE documentrevision DROP CONSTRAINT documentrevision_pkey;
ALTER TABLE documentrevision ADD CONSTRAINT documentrevision_pkey PRIMARY KEY(uuid,revision);

ALTER TABLE documentrevision ALTER COLUMN updatedat DROP NOT NULL;


alter table document add column revision integer;

UPDATE
  document AS t1
SET
  revision = lastrev
FROM (
  SELECT uuid,MAX(revision) as lastrev FROM documentrevision group by "uuid"
) as t2
WHERE 
  t1.uuid = t2.uuid
;
ALTER TABLE document ALTER COLUMN revision SET NOT NULL;
